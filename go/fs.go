package winescape

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

// Standard POSIX / Linux open and access flags.
const (
	O_RDONLY    = 0x0000
	O_WRONLY    = 0x0001
	O_RDWR      = 0x0002
	O_CREAT     = 0x0040
	O_EXCL      = 0x0080
	O_NOCTTY    = 0x0100
	O_TRUNC     = 0x0200
	O_APPEND    = 0x0400
	O_NONBLOCK  = 0x0800
	O_DIRECTORY = 0x10000
	O_CLOEXEC   = 0x80000

	AT_FDCWD            = -100
	AT_SYMLINK_NOFOLLOW = 0x100
	AT_REMOVEDIR        = 0x200
	AT_EACCESS          = 0x200

	SEEK_SET = 0
	SEEK_CUR = 1
	SEEK_END = 2

	DT_UNKNOWN uint8 = 0
	DT_FIFO    uint8 = 1
	DT_CHR     uint8 = 2
	DT_DIR     uint8 = 4
	DT_BLK     uint8 = 6
	DT_REG     uint8 = 8
	DT_LNK     uint8 = 10
	DT_SOCK    uint8 = 12
	DT_WHT     uint8 = 14
)

// Timespec matches the standard 64-bit POSIX struct timespec.
type Timespec struct {
	Sec  int64
	Nsec int64
}

// Stat_t matches the standard 64-bit Linux struct stat layout (144 bytes).
type Stat_t struct {
	Dev     uint64
	Ino     uint64
	Nlink   uint64
	Mode    uint32
	Uid     uint32
	Gid     uint32
	Pad0    uint32
	Rdev    uint64
	Size    int64
	Blksize int64
	Blocks  int64
	Atim    Timespec
	Mtim    Timespec
	Ctim    Timespec
	Pad1    [3]int64
}

// Dirent represents an entry read from getdents64.
type Dirent struct {
	Ino  uint64
	Off  int64
	Type uint8
	Name string
}

// BytePtrFromString returns a pointer to a NUL-terminated byte array, or EINVAL if s contains a NUL byte.
func BytePtrFromString(s string) (*byte, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return nil, syscall.EINVAL
		}
	}
	b := make([]byte, len(s)+1)
	copy(b, s)
	b[len(s)] = 0
	return &b[0], nil
}
// ToUnixPath normalizes DOS/Wine paths (e.g. "Z:\foo\bar", "\\?\unix\foo\bar", "\??\unix\foo\bar")
// into canonical UNIX paths ("/foo/bar") via pure string operations without syscall overhead.
func ToUnixPath(path string) string {
	if path == "" {
		return ""
	}
	p := path
	// Strip \\?\unix\ or \??\unix\ or \\?\unix/ prefixes
	if len(p) >= 9 && (p[:8] == `\\?\unix` || p[:8] == `\??\unix`) && (p[8] == '\\' || p[8] == '/') {
		p = p[8:]
	} else if len(p) >= 3 && (p[0] == 'Z' || p[0] == 'z') && p[1] == ':' && (p[2] == '\\' || p[2] == '/') {
		// Strip Wine's default root drive Z:
		p = p[2:]
	}

	// Normalize backslashes to forward slashes
	b := []byte(p)
	for i := 0; i < len(b); i++ {
		if b[i] == '\\' {
			b[i] = '/'
		}
	}
	res := string(b)
	if len(res) > 0 && res[0] != '/' && (len(path) > 0 && (path[0] == '/' || path[0] == '\\' || (len(path) >= 2 && path[1] == ':'))) {
		res = "/" + res
	}
	return res
}

// IsDir reports whether the file is a directory.
func (s *Stat_t) IsDir() bool {
	return (s.Mode & 0170000) == 0040000 // S_IFDIR
}

// IsRegular reports whether the file is a regular file.
func (s *Stat_t) IsRegular() bool {
	return (s.Mode & 0170000) == 0100000 // S_IFREG
}

// IsSymlink reports whether the file is a symbolic link.
func (s *Stat_t) IsSymlink() bool {
	return (s.Mode & 0170000) == 0120000 // S_IFLNK
}

// Permissions returns the standard POSIX permission bits (07777).
func (s *Stat_t) Permissions() uint32 {
	return s.Mode & 07777
}

// File represents an open file descriptor with standard Go I/O interfaces.
type File struct {
	fd   int
	name string
}

// OpenFile opens a file returning a high-level *File wrapper.
func OpenFile(path string, flags int, mode uint32) (*File, error) {
	fd, err := Open(ToUnixPath(path), flags, mode)
	if err != nil {
		return nil, err
	}
	return &File{fd: fd, name: path}, nil
}

// Fd returns the underlying host integer file descriptor.
func (f *File) Fd() int { return f.fd }

// Name returns the path of the opened file.
func (f *File) Name() string { return f.name }

// Close closes the file descriptor.
func (f *File) Close() error {
	return Close(f.fd)
}

// Read reads up to len(p) bytes from the file.
func (f *File) Read(p []byte) (int, error) {
	return Read(f.fd, p)
}

// Write writes len(p) bytes to the file.
func (f *File) Write(p []byte) (int, error) {
	return Write(f.fd, p)
}

// Seek sets the file offset for the next Read or Write.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	return Seek(f.fd, offset, whence)
}

// ReadAt reads len(p) bytes at specified offset without modifying the file offset.
func (f *File) ReadAt(p []byte, off int64) (int, error) {
	return Pread(f.fd, p, off)
}

// WriteAt writes len(p) bytes at specified offset without modifying the file offset.
func (f *File) WriteAt(p []byte, off int64) (int, error) {
	return Pwrite(f.fd, p, off)
}

// Stat retrieves the file's metadata.
func (f *File) Stat() (*Stat_t, error) {
	var st Stat_t
	if err := Fstat(f.fd, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// ReadDir reads the directory named by path and returns all directory entries using chunked getdents64.
func ReadDir(path string) ([]Dirent, error) {
	fd, err := Open(ToUnixPath(path), O_RDONLY|O_DIRECTORY|O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer Close(fd)

	var allEntries []Dirent
	buf := make([]byte, 65536)
	for {
		n, err := Getdents64(fd, buf)
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			break
		}
		entries, err := ParseDirent64(buf[:n])
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, entries...)
	}
	return allEntries, nil
}

// Getpid returns the process ID of the calling process directly from the host kernel.
func Getpid() (int, error) {
	r1, _, err := Syscall(sysGetpid, 0, 0, 0)
	if err != nil {
		return 0, err
	}
	return int(r1), nil
}

// Openat opens a file relative to dirfd.
func Openat(dirfd int, path string, flags int, mode uint32) (int, error) {
	p, err := BytePtrFromString(path)
	if err != nil {
		return -1, err
	}
	r1, _, errSys := Syscall6(sysOpenat, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(flags), uintptr(mode), 0, 0)
	if errSys != nil {
		return -1, errSys
	}
	return int(r1), nil
}

// Open opens a file at path.
func Open(path string, flags int, mode uint32) (int, error) {
	return Openat(AT_FDCWD, path, flags, mode)
}

// Close closes an open file descriptor.
func Close(fd int) error {
	_, _, err := Syscall(sysClose, uintptr(fd), 0, 0)
	return err
}

// Read reads up to len(p) bytes into p from fd.
func Read(fd int, p []byte) (int, error) {
	var ptr uintptr
	if len(p) > 0 {
		ptr = uintptr(unsafe.Pointer(&p[0]))
	}
	r1, _, err := Syscall(sysRead, uintptr(fd), ptr, uintptr(len(p)))
	if err != nil {
		return 0, err
	}
	return int(r1), nil
}

// Write writes len(p) bytes from p to fd.
func Write(fd int, p []byte) (int, error) {
	var ptr uintptr
	if len(p) > 0 {
		ptr = uintptr(unsafe.Pointer(&p[0]))
	}
	r1, _, err := Syscall(sysWrite, uintptr(fd), ptr, uintptr(len(p)))
	if err != nil {
		return 0, err
	}
	return int(r1), nil
}

// Pread reads len(p) bytes into p from fd at offset without modifying the file offset.
func Pread(fd int, p []byte, offset int64) (int, error) {
	var ptr uintptr
	if len(p) > 0 {
		ptr = uintptr(unsafe.Pointer(&p[0]))
	}
	r1, _, err := Syscall6(sysPread64, uintptr(fd), ptr, uintptr(len(p)), uintptr(offset), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(r1), nil
}

// Pwrite writes len(p) bytes from p to fd at offset without modifying the file offset.
func Pwrite(fd int, p []byte, offset int64) (int, error) {
	var ptr uintptr
	if len(p) > 0 {
		ptr = uintptr(unsafe.Pointer(&p[0]))
	}
	r1, _, err := Syscall6(sysPwrite64, uintptr(fd), ptr, uintptr(len(p)), uintptr(offset), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(r1), nil
}

// Seek sets the offset for the next Read or Write on fd.
func Seek(fd int, offset int64, whence int) (int64, error) {
	r1, _, err := Syscall(sysLseek, uintptr(fd), uintptr(offset), uintptr(whence))
	if err != nil {
		return 0, err
	}
	return int64(r1), nil
}

// Fstat retrieves file status for open fd.
func Fstat(fd int, stat *Stat_t) error {
	_, _, err := Syscall(sysFstat, uintptr(fd), uintptr(unsafe.Pointer(stat)), 0)
	return err
}

// Fstatat retrieves file status relative to dirfd.
func Fstatat(dirfd int, path string, stat *Stat_t, flags int) error {
	p, err := BytePtrFromString(path)
	if err != nil {
		return err
	}
	var nr uintptr = sysNewfstatat
	if nr == 0 {
		nr = sysFstatat
	}
	_, _, errSys := Syscall6(nr, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(stat)), uintptr(flags), 0, 0)
	return errSys
}

// Stat retrieves file status for path.
func Stat(path string, stat *Stat_t) error {
	return Fstatat(AT_FDCWD, path, stat, 0)
}

// Lstat retrieves file status for path without following symlinks.
func Lstat(path string, stat *Stat_t) error {
	return Fstatat(AT_FDCWD, path, stat, AT_SYMLINK_NOFOLLOW)
}

// Unlinkat removes a directory entry relative to dirfd.
func Unlinkat(dirfd int, path string, flags int) error {
	p, err := BytePtrFromString(path)
	if err != nil {
		return err
	}
	_, _, errSys := Syscall(sysUnlinkat, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(flags))
	return errSys
}

// Unlink removes a file.
func Unlink(path string) error {
	return Unlinkat(AT_FDCWD, path, 0)
}

// Rmdir removes an empty directory.
func Rmdir(path string) error {
	return Unlinkat(AT_FDCWD, path, AT_REMOVEDIR)
}

// Mkdirat creates a directory relative to dirfd.
func Mkdirat(dirfd int, path string, mode uint32) error {
	p, err := BytePtrFromString(path)
	if err != nil {
		return err
	}
	_, _, errSys := Syscall(sysMkdirat, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(mode))
	return errSys
}

// Mkdir creates a directory at path.
func Mkdir(path string, mode uint32) error {
	return Mkdirat(AT_FDCWD, path, mode)
}

// Renameat renames a file relative to directory file descriptors.
func Renameat(olddirfd int, oldpath string, newdirfd int, newpath string) error {
	oldP, err := BytePtrFromString(oldpath)
	if err != nil {
		return err
	}
	newP, err := BytePtrFromString(newpath)
	if err != nil {
		return err
	}
	_, _, errSys := Syscall6(sysRenameat, uintptr(olddirfd), uintptr(unsafe.Pointer(oldP)), uintptr(newdirfd), uintptr(unsafe.Pointer(newP)), 0, 0)
	return errSys
}

// Rename renames oldpath to newpath.
func Rename(oldpath, newpath string) error {
	return Renameat(AT_FDCWD, oldpath, AT_FDCWD, newpath)
}

// Readlinkat reads the value of a symbolic link relative to dirfd.
func Readlinkat(dirfd int, path string, buf []byte) (int, error) {
	p, err := BytePtrFromString(path)
	if err != nil {
		return 0, err
	}
	var ptr uintptr
	if len(buf) > 0 {
		ptr = uintptr(unsafe.Pointer(&buf[0]))
	}
	r1, _, errSys := Syscall6(sysReadlinkat, uintptr(dirfd), uintptr(unsafe.Pointer(p)), ptr, uintptr(len(buf)), 0, 0)
	if errSys != nil {
		return 0, errSys
	}
	return int(r1), nil
}

// Readlink reads the value of a symbolic link.
func Readlink(path string, buf []byte) (int, error) {
	return Readlinkat(AT_FDCWD, path, buf)
}

// Faccessat checks file accessibility relative to dirfd.
func Faccessat(dirfd int, path string, mode uint32, flags int) error {
	p, err := BytePtrFromString(path)
	if err != nil {
		return err
	}
	_, _, errSys := Syscall6(sysFaccessat, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(mode), uintptr(flags), 0, 0)
	return errSys
}

// Access checks file accessibility at path.
func Access(path string, mode uint32) error {
	return Faccessat(AT_FDCWD, path, mode, 0)
}

// Getdents64 reads directory entries from fd into buf.
func Getdents64(fd int, buf []byte) (int, error) {
	var ptr uintptr
	if len(buf) > 0 {
		ptr = uintptr(unsafe.Pointer(&buf[0]))
	}
	var nr uintptr = sysGetdents64
	if nr == 0 {
		nr = sysGetdirentries
	}
	r1, _, err := Syscall(nr, uintptr(fd), ptr, uintptr(len(buf)))
	if err != nil {
		return 0, err
	}
	return int(r1), nil
}

// ParseDirent64 parses entries from raw getdents64 buffer.
func ParseDirent64(buf []byte) ([]Dirent, error) {
	var entries []Dirent
	cur := 0
	for cur+19 <= len(buf) {
		ino := binary.LittleEndian.Uint64(buf[cur : cur+8])
		off := int64(binary.LittleEndian.Uint64(buf[cur+8 : cur+16]))
		reclen := binary.LittleEndian.Uint16(buf[cur+16 : cur+18])
		if reclen < 19 || cur+int(reclen) > len(buf) {
			break
		}
		dtype := buf[cur+18]

		nameBytes := buf[cur+19 : cur+int(reclen)]
		nameEnd := 0
		for nameEnd < len(nameBytes) && nameBytes[nameEnd] != 0 {
			nameEnd++
		}

		name := string(nameBytes[:nameEnd])
		if ino != 0 {
			entries = append(entries, Dirent{
				Ino:  ino,
				Off:  off,
				Type: dtype,
				Name: name,
			})
		}
		cur += int(reclen)
	}
	return entries, nil
}
