//go:build windows

package winescape

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	isWineCached     bool
	isWineChecked    bool
	hostOSCached     string
	hostOSChecked    bool
	availableCached  bool
	availableChecked bool
	selfTestErr      error
)

// IsWine returns true if running under the Wine compatibility layer.
func IsWine() bool {
	if isWineChecked {
		return isWineCached
	}
	ntdll, err := syscall.LoadDLL("ntdll.dll")
	if err == nil {
		_, err := ntdll.FindProc("wine_get_version")
		isWineCached = (err == nil)
	}
	isWineChecked = true
	return isWineCached
}

// HostOS returns the host operating system under Wine ("linux", "freebsd", "darwin", etc.), or "" on native Windows.
func HostOS() string {
	if hostOSChecked {
		return hostOSCached
	}
	if !IsWine() {
		hostOSChecked = true
		return ""
	}
	ntdll, err := syscall.LoadDLL("ntdll.dll")
	if err == nil {
		proc, err := ntdll.FindProc("wine_get_host_version")
		if err == nil {
			var sysname *byte
			var release *byte
			proc.Call(uintptr(unsafe.Pointer(&sysname)), uintptr(unsafe.Pointer(&release)))
			if sysname != nil {
				hostOSCached = gostring(sysname)
			}
		}
	}
	if hostOSCached == "" {
		hostOSCached = "linux"
	}
	hostOSChecked = true
	return hostOSCached
}

func gostring(p *byte) string {
	if p == nil {
		return ""
	}
	var b []byte
	for *p != 0 {
		b = append(b, *p)
		p = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + 1))
	}
	return string(b)
}

// Available returns true if raw host syscalls are supported in the current runtime environment.
//
// Restricted to Linux hosts only. The generic Go trampoline (syscall6_raw,
// go/trampoline_windows_amd64.s) implements exactly one error-signalling
// convention: a negative two's-complement value in AX means -errno, which is
// the Linux calling convention. FreeBSD signals errors via the carry flag
// with a *positive* errno in AX instead (see c/src/trampoline_freebsd_amd64.S,
// which already implements this correctly for C consumers) -- the Go side has
// no equivalent trampoline yet. Advertising FreeBSD support here before that
// exists would make every Go caller silently misinterpret both FreeBSD
// success values (as spurious errors, for large results) and FreeBSD errors
// (as success, since the real error path never triggers), against wrong
// syscall numbers to boot: the numbers_freebsd_amd64.go table additionally
// requires the winescape_freebsd build tag, which is not on by default, so a
// plain build reaching this codepath would even dispatch FreeBSD's numbers
// as if they were Linux's.
//
// TODO(libwinescape): add a carry-flag-aware syscall6_raw variant for
// windows/amd64 (mirroring c/src/trampoline_freebsd_amd64.S), a matching
// windows/arm64 one if BSD-arm64 is ever probed (see spec/table.go), wire
// Syscall6/RawSyscall6 to pick the right trampoline based on HostOS(), and
// only then re-enable "freebsd" here.
func Available() bool {
	if availableChecked {
		return availableCached
	}
	if !IsWine() {
		availableChecked = true
		return false
	}
	os := strings.ToLower(HostOS())
	// Raw kernel traps from PE code are supported today only on Linux hosts;
	// see the TODO above for why FreeBSD is deliberately not included yet.
	if os != "linux" && os != "gnu/linux" {
		availableCached = false
		availableChecked = true
		return false
	}
	// Belt-and-braces against the standing risk docs/RISK-REVIEW-farmanager-1123.md
	// item 4 accepts as real: nothing guarantees Wine's raw-syscall passthrough
	// stays behaved in a future version. A supported HostOS() alone only means
	// "this used to work here" -- selfTest() actually exercises the code path
	// this library's callers will use (open/write/read/close/unlink a real
	// file, plus one deliberate-failure path) right now, on this machine, on
	// this Wine version, before anything trusts it. See selfTest's own doc
	// comment for exactly what this does and doesn't catch.
	if err := selfTest(); err != nil {
		selfTestErr = err
		availableCached = false
		availableChecked = true
		return false
	}
	availableCached = true
	availableChecked = true
	return true
}

// SelfTestError returns the reason Available() returned false due to a
// failed startup self-test, or nil if Available() hasn't been called yet,
// returned true, or failed for a reason other than the self-test (unsupported
// host OS, not running under Wine). Exists so a caller that wants to log or
// display *why* the fast path isn't available can do so instead of only
// knowing that it isn't.
func SelfTestError() error {
	return selfTestErr
}
