package winescape

import (
	"unsafe"
)

// Linux/POSIX ioctl and terminal constants.
const (
	TIOCGWINSZ = 0x5413
	TIOCSWINSZ = 0x5414

	CLOCK_REALTIME  = 0
	CLOCK_MONOTONIC = 1
	CLOCK_BOOTTIME  = 7

	SIGINT   = 2
	SIGKILL  = 9
	SIGTERM  = 15
	SIGCHLD  = 17
	SIGWINCH = 28
)

// Winsize matches struct winsize in Linux/POSIX (8 bytes).
type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// Ioctl performs a generic device control operation on file descriptor fd.
func Ioctl(fd int, req uintptr, arg unsafe.Pointer) error {
	_, _, err := Syscall(sysIoctl, uintptr(fd), req, uintptr(arg))
	return err
}

// GetWinsize returns the terminal window dimensions of fd via direct TIOCGWINSZ ioctl.
func GetWinsize(fd int) (*Winsize, error) {
	var ws Winsize
	if err := Ioctl(fd, TIOCGWINSZ, unsafe.Pointer(&ws)); err != nil {
		return nil, err
	}
	return &ws, nil
}

// ClockGettime retrieves the current time from the specified POSIX clock directly from the kernel.
func ClockGettime(clockid int, ts *Timespec) error {
	_, _, err := Syscall(sysClockGettime, uintptr(clockid), uintptr(unsafe.Pointer(ts)), 0)
	return err
}

// ClockNanosleep pauses thread execution for the specified duration using the specified POSIX clock.
func ClockNanosleep(clockid int, flags int, req *Timespec, rem *Timespec) error {
	var remPtr uintptr
	if rem != nil {
		remPtr = uintptr(unsafe.Pointer(rem))
	}
	var nr uintptr = sysClockNanosleep
	if nr == 0 {
		nr = sysNanosleep
		_, _, err := Syscall(nr, uintptr(unsafe.Pointer(req)), remPtr, 0)
		return err
	}
	_, _, err := Syscall6(nr, uintptr(clockid), uintptr(flags), uintptr(unsafe.Pointer(req)), remPtr, 0, 0)
	return err
}

// Nanosleep pauses thread execution for the specified duration using CLOCK_REALTIME.
func Nanosleep(req *Timespec, rem *Timespec) error {
	return ClockNanosleep(CLOCK_REALTIME, 0, req, rem)
}

// Kill sends signal sig to process pid directly via the host kernel.
func Kill(pid int, sig int) error {
	_, _, err := Syscall(sysKill, uintptr(pid), uintptr(sig), 0)
	return err
}

// Wait4 waits for process state changes on pid.
func Wait4(pid int, status *int32, options int, rusage unsafe.Pointer) (int, error) {
	var statPtr uintptr
	if status != nil {
		statPtr = uintptr(unsafe.Pointer(status))
	}
	r1, _, err := Syscall6(sysWait4, uintptr(pid), statPtr, uintptr(options), uintptr(rusage), 0, 0)
	if err != nil {
		return -1, err
	}
	return int(r1), nil
}
