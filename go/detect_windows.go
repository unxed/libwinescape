//go:build windows

package winescape

import (
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
func Available() bool {
	if availableChecked {
		return availableCached
	}
	if !IsWine() {
		availableChecked = true
		return false
	}
	os := HostOS()
	// Raw kernel traps from PE code are supported on Linux and FreeBSD hosts
	availableCached = (os == "linux" || os == "freebsd" || os == "GNU/Linux")
	availableChecked = true
	return availableCached
}
