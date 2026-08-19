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

// Standard POSIX / Linux errno constants.
const (
	EPERM    = Errno(1)
	ENOENT   = Errno(2)
	ESRCH    = Errno(3)
	EINTR    = Errno(4)
	EIO      = Errno(5)
	EBADF    = Errno(9)
	ECHILD   = Errno(10)
	EAGAIN   = Errno(11)
	ENOMEM   = Errno(12)
	EACCES   = Errno(13)
	EFAULT   = Errno(14)
	EBUSY    = Errno(16)
	EEXIST   = Errno(17)
	ENODEV   = Errno(19)
	ENOTDIR  = Errno(20)
	EISDIR   = Errno(21)
	EINVAL   = Errno(22)
	ENFILE   = Errno(23)
	EMFILE   = Errno(24)
	ENOTTY   = Errno(25)
	ETXTBSY  = Errno(26)
	ESPIPE   = Errno(29)
	EPIPE    = Errno(32)
	ENOSYS   = Errno(38)
	ERESTART = Errno(516)
)

var (
	// ErrUnavailable is returned when calling raw syscalls on native Windows or an unsupported host OS.
	ErrUnavailable = syscall.ENOSYS
)

// Syscall6 issues a raw 6-argument system call to the underlying host kernel.
func Syscall6(nr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err error) {
	if !Available() {
		return 0, 0, ErrUnavailable
	}
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

// SyscallN issues a raw system call with 0 to 6 variable arguments.
func SyscallN(nr uintptr, args ...uintptr) (r1, r2 uintptr, err error) {
	var a [6]uintptr
	for i := 0; i < len(args) && i < 6; i++ {
		a[i] = args[i]
	}
	return Syscall6(nr, a[0], a[1], a[2], a[3], a[4], a[5])
}

// Call executes a raw syscall with 0 to 6 variable arguments, returning (result, error).
func Call(nr uintptr, args ...uintptr) (uintptr, error) {
	r1, _, err := SyscallN(nr, args...)
	return r1, err
}

// retryEINTR re-issues a raw syscall while it fails with EINTR or ERESTART.
//
// Because Syscall6/Syscall bypass the Go runtime's entersyscall/exitsyscall
// hooks, none of the usual retry-on-EINTR machinery that the standard
// "syscall" package provides (see internal/poll.ignoringEINTR) applies here.
// The host kernel still delivers SIGURG (Go's async preemption signal) and,
// if a profiler is active, SIGPROF to the OS thread at any time; the Go
// runtime installs all of its signal handlers with SA_RESTART, so most
// interrupted blocking calls are silently restarted by the kernel itself,
// but a few classes of syscalls are documented to never auto-restart
// regardless of SA_RESTART (see signal(7)): most notably connect(2), and
// read/recv/accept-family calls on a socket that has an SO_RCVTIMEO/
// SO_SNDTIMEO timeout set. retryEINTR is the userspace fallback for those.
//
// It must ONLY be used to wrap calls that, per POSIX, have no observable
// side effect when interrupted before completing (a fresh attempt is
// equivalent to the original one) — e.g. read, write, open, flock, wait4,
// accept4. It must NEVER be used for close(2) (a signal arriving after the
// descriptor was already released, but before the raw syscall returned to
// Go, would make a blind retry close an unrelated, meanwhile-reused fd) or
// for connect(2) on a blocking socket (POSIX requires waiting for
// writability via select/poll and checking SO_ERROR instead of calling
// connect again; see the comment on Connect).
func retryEINTR(fn func() (uintptr, uintptr, error)) (uintptr, uintptr, error) {
	for {
		r1, r2, err := fn()
		if err == EINTR || err == ERESTART {
			continue
		}
		return r1, r2, err
	}
}
