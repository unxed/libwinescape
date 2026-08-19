package winescape

import (
	"syscall"
	"unsafe"
)

// Socket domain and type constants.
const (
	AF_UNIX  = 1
	AF_LOCAL = 1
	AF_INET  = 2
	AF_INET6 = 10

	SOCK_STREAM   = 1
	SOCK_DGRAM    = 2
	SOCK_RAW      = 3
	SOCK_NONBLOCK = 0x00000800
	SOCK_CLOEXEC  = 0x00080000

	SHUT_RD   = 0
	SHUT_WR   = 1
	SHUT_RDWR = 2
)

// RawSockaddrUn matches struct sockaddr_un in POSIX (110 bytes).
type RawSockaddrUn struct {
	Family uint16
	Path   [108]byte
}

// Socket creates a kernel socket.
func Socket(domain, typ, proto int) (int, error) {
	r1, _, err := Syscall(sysSocket, uintptr(domain), uintptr(typ), uintptr(proto))
	if err != nil {
		return -1, err
	}
	return int(r1), nil
}

// Connect connects socket fd to the address specified by addr.
//
// Unlike Read/Write/Accept4, this deliberately does NOT retry on EINTR.
// connect(2) is explicitly called out by signal(7) as never automatically
// restarted, and — more importantly — it is not safe to simply call it
// again: once the kernel has begun establishing the connection, a second
// connect() on the same fd fails with EALREADY (or EISCONN if it already
// succeeded), not a clean retry. If Connect returns EINTR, the correct
// recovery is for the caller to poll fd for writability (e.g. via
// select/poll) and then check SO_ERROR, exactly as documented in
// connect(2); this package does not currently provide that helper.
func Connect(fd int, addr unsafe.Pointer, addrlen uintptr) error {
	_, _, err := Syscall(sysConnect, uintptr(fd), uintptr(addr), addrlen)
	return err
}

// Bind binds socket fd to the address specified by addr.
func Bind(fd int, addr unsafe.Pointer, addrlen uintptr) error {
	_, _, err := Syscall(sysBind, uintptr(fd), uintptr(addr), addrlen)
	return err
}

// Listen sets socket fd to listening mode with given backlog.
func Listen(fd int, backlog int) error {
	_, _, err := Syscall(sysListen, uintptr(fd), uintptr(backlog), 0)
	return err
}

// Accept4 accepts a connection on socket fd with flags (e.g. SOCK_CLOEXEC, SOCK_NONBLOCK).
func Accept4(fd int, addr unsafe.Pointer, addrlen *uint32, flags int) (int, error) {
	var addrPtr, lenPtr uintptr
	if addr != nil {
		addrPtr = uintptr(addr)
	}
	if addrlen != nil {
		lenPtr = uintptr(unsafe.Pointer(addrlen))
	}
	// accept4() interrupted before a pending connection is dequeued has no
	// side effect (the connection, if any, stays queued for the next call),
	// so it's safe to retry — unlike Connect.
	r1, _, err := retryEINTR(func() (uintptr, uintptr, error) {
		return Syscall6(sysAccept4, uintptr(fd), addrPtr, lenPtr, uintptr(flags), 0, 0)
	})
	if err != nil {
		return -1, err
	}
	return int(r1), nil
}

// DialUnix connects directly to a host AF_UNIX socket (X11, Wayland, D-Bus, Docker, ssh-agent)
// using raw kernel syscalls without relying on Windows Winsock AF_UNIX support.
func DialUnix(path string) (*File, error) {
	uPath := ToUnixPath(path)
	if len(uPath) == 0 || len(uPath) >= 108 {
		return nil, syscall.EINVAL
	}

	fd, err := Socket(AF_UNIX, SOCK_STREAM|SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	var sa RawSockaddrUn
	sa.Family = AF_UNIX
	copy(sa.Path[:], uPath)
	saLen := uintptr(2 + len(uPath) + 1)

	if err := Connect(fd, unsafe.Pointer(&sa), saLen); err != nil {
		Close(fd)
		return nil, err
	}

	return &File{fd: fd, name: uPath}, nil
}

// ListenUnix creates an AF_UNIX listening socket bound to path.
func ListenUnix(path string, backlog int) (int, error) {
	uPath := ToUnixPath(path)
	if len(uPath) == 0 || len(uPath) >= 108 {
		return -1, syscall.EINVAL
	}

	fd, err := Socket(AF_UNIX, SOCK_STREAM|SOCK_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}

	var sa RawSockaddrUn
	sa.Family = AF_UNIX
	copy(sa.Path[:], uPath)
	saLen := uintptr(2 + len(uPath) + 1)

	if err := Bind(fd, unsafe.Pointer(&sa), saLen); err != nil {
		Close(fd)
		return -1, err
	}

	if backlog <= 0 {
		backlog = 128
	}
	if err := Listen(fd, backlog); err != nil {
		Close(fd)
		return -1, err
	}

	return fd, nil
}
