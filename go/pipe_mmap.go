package winescape

import (
	"unsafe"
)

// POSIX mmap protection and flags constants.
const (
	PROT_NONE  = 0x0
	PROT_READ  = 0x1
	PROT_WRITE = 0x2
	PROT_EXEC  = 0x4

	MAP_SHARED    = 0x01
	MAP_PRIVATE   = 0x02
	MAP_ANONYMOUS = 0x20
)

// Pipe2 creates an anonymous pipe with flags (e.g. O_CLOEXEC, O_NONBLOCK).
func Pipe2(flags int) (r, w int, err error) {
	var p [2]int32
	_, _, errSys := Syscall(sysPipe2, uintptr(unsafe.Pointer(&p[0])), uintptr(flags), 0)
	if errSys != nil {
		return -1, -1, errSys
	}
	return int(p[0]), int(p[1]), nil
}

// Dup3 duplicates oldfd to newfd with flags (e.g. O_CLOEXEC).
func Dup3(oldfd, newfd, flags int) error {
	_, _, err := Syscall(sysDup3, uintptr(oldfd), uintptr(newfd), uintptr(flags))
	return err
}

// Mmap maps length bytes of file fd starting at offset into memory.
func Mmap(fd int, offset int64, length int, prot, flags int) ([]byte, error) {
	r1, _, err := Syscall6(sysMmap, 0, uintptr(length), uintptr(prot), uintptr(flags), uintptr(fd), uintptr(offset))
	if err != nil {
		return nil, err
	}
	var b []byte
	hdr := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(&b))
	hdr.Data = r1
	hdr.Len = length
	hdr.Cap = length
	return b, nil
}

// Munmap unmaps a previously mapped memory region.
func Munmap(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	_, _, err := Syscall(sysMunmap, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), 0)
	return err
}
