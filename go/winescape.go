package winescape

import "syscall"

// Errno represents a host kernel error number.
type Errno uintptr

func (e Errno) Error() string {
	if e == 0 {
		return "errno 0 (success)"
	}
	return syscall.Errno(e).Error()
}

// Syscall6 issues a raw 6-argument system call to the underlying host kernel.
func Syscall6(nr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err error) {
	r1, r2, errno := syscall6_raw(nr, a1, a2, a3, a4, a5, a6)
	if errno != 0 {
		return r1, r2, Errno(errno)
	}
	return r1, r2, nil
}

// Syscall issues a raw system call with up to 3 arguments.
func Syscall(nr, a1, a2, a3 uintptr) (r1, r2 uintptr, err error) {
	return Syscall6(nr, a1, a2, a3, 0, 0, 0)
}

// RawSyscall6 is identical to Syscall6 but returns Errno directly without error wrapping.
func RawSyscall6(nr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno) {
	r1, r2, errno := syscall6_raw(nr, a1, a2, a3, a4, a5, a6)
	return r1, r2, Errno(errno)
}

// RawSyscall is identical to Syscall but returns Errno directly.
func RawSyscall(nr, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno) {
	return RawSyscall6(nr, a1, a2, a3, 0, 0, 0)
}
