package winescape

import (
	"io/fs"
	"testing"
	"time"
)

func TestFileInfo_Interface(t *testing.T) {
	fi := &FileInfo{
		name: "test.txt",
		st: Stat_t{
			Size: 1024,
			Mode: 0100644,
			Mtim: Timespec{Sec: 1700000000, Nsec: 500000000},
		},
	}

	var _ fs.FileInfo = fi

	if fi.Name() != "test.txt" {
		t.Errorf("expected Name 'test.txt', got %q", fi.Name())
	}
	if fi.Size() != 1024 {
		t.Errorf("expected Size 1024, got %d", fi.Size())
	}
	if fi.IsDir() {
		t.Errorf("expected IsDir false")
	}
	if fi.Mode() != 0644 {
		t.Errorf("expected Mode 0644, got %v", fi.Mode())
	}
	expectedTime := time.Unix(1700000000, 500000000)
	if !fi.ModTime().Equal(expectedTime) {
		t.Errorf("expected ModTime %v, got %v", expectedTime, fi.ModTime())
	}
}

func TestDirEntry_Interface(t *testing.T) {
	de := &DirEntry{
		d: Dirent{
			Name: "subdir",
			Type: DT_DIR,
		},
		dir: "/tmp",
	}

	var _ fs.DirEntry = de

	if de.Name() != "subdir" {
		t.Errorf("expected Name 'subdir', got %q", de.Name())
	}
	if !de.IsDir() {
		t.Errorf("expected IsDir true")
	}
	if de.Type() != fs.ModeDir {
		t.Errorf("expected Type ModeDir, got %v", de.Type())
	}
}
