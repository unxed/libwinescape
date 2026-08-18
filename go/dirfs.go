package winescape

import (
	"io"
	"io/fs"
	"path"
	"time"
)

// FileInfo wraps Stat_t to implement the standard io/fs.FileInfo interface.
type FileInfo struct {
	name string
	st   Stat_t
}

func (fi *FileInfo) Name() string       { return fi.name }
func (fi *FileInfo) Size() int64        { return fi.st.Size }
func (fi *FileInfo) ModTime() time.Time { return time.Unix(fi.st.Mtim.Sec, fi.st.Mtim.Nsec) }
func (fi *FileInfo) IsDir() bool        { return fi.st.IsDir() }
func (fi *FileInfo) Sys() any           { return &fi.st }
func (fi *FileInfo) Mode() fs.FileMode {
	m := fs.FileMode(fi.st.Permissions())
	if fi.st.IsDir() {
		m |= fs.ModeDir
	}
	if fi.st.IsSymlink() {
		m |= fs.ModeSymlink
	}
	return m
}

// DirEntry wraps Dirent to implement the standard io/fs.DirEntry interface.
type DirEntry struct {
	d   Dirent
	dir string
}

func (de *DirEntry) Name() string { return de.d.Name }
func (de *DirEntry) IsDir() bool  { return de.d.Type == DT_DIR }
func (de *DirEntry) Type() fs.FileMode {
	switch de.d.Type {
	case DT_DIR:
		return fs.ModeDir
	case DT_LNK:
		return fs.ModeSymlink
	case DT_FIFO:
		return fs.ModeNamedPipe
	case DT_SOCK:
		return fs.ModeSocket
	case DT_CHR:
		return fs.ModeDevice | fs.ModeCharDevice
	case DT_BLK:
		return fs.ModeDevice
	default:
		return 0
	}
}

func (de *DirEntry) Info() (fs.FileInfo, error) {
	fullPath := path.Join(de.dir, de.d.Name)
	var st Stat_t
	if err := Lstat(fullPath, &st); err != nil {
		return nil, err
	}
	return &FileInfo{name: de.d.Name, st: st}, nil
}

// DirFS returns a standard io/fs.FS rooted at the specified host UNIX directory.
func DirFS(dir string) fs.FS {
	return &dirFS{root: ToUnixPath(dir)}
}

type dirFS struct {
	root string
}

func (d *dirFS) join(name string) string {
	if !fs.ValidPath(name) {
		return ""
	}
	if name == "." {
		return d.root
	}
	return path.Join(d.root, name)
}

func (d *dirFS) Open(name string) (fs.File, error) {
	fullPath := d.join(name)
	if fullPath == "" {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	file, err := OpenFile(fullPath, O_RDONLY|O_CLOEXEC, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	return &openDirFile{File: file, name: path.Base(name), fullPath: fullPath}, nil
}

func (d *dirFS) Stat(name string) (fs.FileInfo, error) {
	fullPath := d.join(name)
	if fullPath == "" {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}
	var st Stat_t
	if err := Stat(fullPath, &st); err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}
	return &FileInfo{name: path.Base(name), st: st}, nil
}

func (d *dirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	path := d.join(name)
	if path == "" {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	entries, err := ReadDir(path)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}
	res := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		res[i] = &DirEntry{d: e, dir: path}
	}
	return res, nil
}

func (d *dirFS) ReadFile(name string) ([]byte, error) {
	path := d.join(name)
	if path == "" {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrInvalid}
	}
	var st Stat_t
	if err := Stat(path, &st); err != nil {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: err}
	}
	f, err := OpenFile(path, O_RDONLY|O_CLOEXEC, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: err}
	}
	defer f.Close()

	buf := make([]byte, st.Size)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: err}
	}
	return buf[:n], nil
}

type openDirFile struct {
	*File
	name     string
	fullPath string
}

func (o *openDirFile) Stat() (fs.FileInfo, error) {
	st, err := o.File.Stat()
	if err != nil {
		return nil, err
	}
	return &FileInfo{name: o.name, st: *st}, nil
}

func (o *openDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	entries, err := ReadDir(o.fullPath)
	if err != nil {
		return nil, err
	}
	res := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		res[i] = &DirEntry{d: e, dir: o.fullPath}
	}
	if n > 0 && n < len(res) {
		return res[:n], nil
	}
	return res, nil
}
