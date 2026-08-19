package winescape

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var tempFileCounter uint64

// ReadFile reads the named file and returns the contents directly via raw kernel syscalls.
func ReadFile(filename string) ([]byte, error) {
	fd, err := Open(filename, O_RDONLY|O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer Close(fd)

	var st Stat_t
	if err := Fstat(fd, &st); err != nil {
		return nil, err
	}

	size := int(st.Size)
	if size < 0 {
		size = 0
	}
	buf := make([]byte, size)
	n, err := Read(fd, buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf[:n], nil
}

// WriteFile writes data to the named file directly via raw kernel syscalls.
func WriteFile(filename string, data []byte, perm uint32) error {
	fd, err := Open(filename, O_WRONLY|O_CREAT|O_TRUNC|O_CLOEXEC, perm)
	if err != nil {
		return err
	}
	defer Close(fd)

	if len(data) > 0 {
		_, err = Write(fd, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents.
func MkdirAll(dirpath string, perm uint32) error {
	uPath := ToUnixPath(dirpath)
	if uPath == "" || uPath == "/" || uPath == "." {
		return nil
	}

	var st Stat_t
	if err := Stat(uPath, &st); err == nil {
		if st.IsDir() {
			return nil
		}
		return syscall.ENOTDIR
	}

	parts := strings.Split(strings.Trim(uPath, "/"), "/")
	current := ""
	if strings.HasPrefix(uPath, "/") {
		current = "/"
	}

	for _, part := range parts {
		if part == "" {
			continue
		}
		if current == "/" {
			current += part
		} else if current == "" {
			current = part
		} else {
			current += "/" + part
		}

		if err := Mkdir(current, perm); err != nil {
			if !errors.Is(err, syscall.EEXIST) && !errors.Is(err, syscall.EISDIR) {
				var stCur Stat_t
				if errStat := Stat(current, &stCur); errStat == nil && stCur.IsDir() {
					continue
				}
				return err
			}
		}
	}
	return nil
}

// RemoveAll removes path and any children it contains directly via fast kernel getdents64 traversal.
func RemoveAll(targetPath string) error {
	uPath := ToUnixPath(targetPath)
	if uPath == "" || uPath == "/" || uPath == "." {
		return errors.New("winescape: refusing to RemoveAll root or empty path")
	}

	var st Stat_t
	if err := Lstat(uPath, &st); err != nil {
		if errors.Is(err, syscall.ENOENT) {
			return nil
		}
		return err
	}

	if !st.IsDir() {
		return Unlink(uPath)
	}

	entries, err := ReadDir(uPath)
	if err == nil {
		for _, e := range entries {
			if e.Name == "." || e.Name == ".." {
				continue
			}
			child := path.Join(uPath, e.Name)
			if errChild := RemoveAll(child); errChild != nil && !errors.Is(errChild, syscall.ENOENT) {
				return errChild
			}
		}
	}

	return Rmdir(uPath)
}

// CopyFile copies src to dst, attempting in-kernel CopyFileRange first, and falling back to a chunked buffer.
func CopyFile(src, dst string) error {
	srcFd, err := Open(src, O_RDONLY|O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer Close(srcFd)

	var st Stat_t
	if err := Fstat(srcFd, &st); err != nil {
		return err
	}

	perm := st.Permissions()
	if perm == 0 {
		perm = 0644
	}

	dstFd, err := Open(dst, O_WRONLY|O_CREAT|O_TRUNC|O_CLOEXEC, perm)
	if err != nil {
		return err
	}
	defer Close(dstFd)

	totalSize := int(st.Size)
	if totalSize > 0 {
		copied, err := CopyFileRange(srcFd, nil, dstFd, nil, totalSize, 0)
		if err == nil && copied == totalSize {
			return nil
		}

		if _, errSeek := Seek(srcFd, 0, SEEK_SET); errSeek != nil {
			return errSeek
		}
		if _, errSeek := Seek(dstFd, 0, SEEK_SET); errSeek != nil {
			return errSeek
		}

		buf := make([]byte, 65536)
		for {
			nr, errRead := Read(srcFd, buf)
			if nr > 0 {
				nw, errWrite := Write(dstFd, buf[:nr])
				if errWrite != nil {
					return errWrite
				}
				if nw != nr {
					return io.ErrShortWrite
				}
			}
			if errRead != nil {
				if errRead == io.EOF {
					break
				}
				return errRead
			}
		}
	}

	return nil
}

// CreateTemp creates a new temporary file in dir with pattern.
func CreateTemp(dir, pattern string) (*File, error) {
	if dir == "" {
		dir = "/tmp"
	}
	dir = ToUnixPath(dir)

	prefix, suffix := pattern, ""
	if pos := strings.LastIndex(pattern, "*"); pos != -1 {
		prefix, suffix = pattern[:pos], pattern[pos+1:]
	}

	for i := 0; i < 1000; i++ {
		count := atomic.AddUint64(&tempFileCounter, 1)
		name := path.Join(dir, fmt.Sprintf("%s%d_%d%s", prefix, time.Now().UnixNano(), count, suffix))
		f, err := OpenFile(name, O_RDWR|O_CREAT|O_EXCL|O_CLOEXEC, 0600)
		if err == nil {
			return f, nil
		}
		if !errors.Is(err, syscall.EEXIST) {
			return nil, err
		}
	}
	return nil, syscall.EEXIST
}
