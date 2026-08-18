package winescape

import (
	"testing"
	"unsafe"
)

func TestStatfs_t_Size(t *testing.T) {
	var st Statfs_t
	sz := unsafe.Sizeof(st)
	if sz != 120 {
		t.Errorf("expected Statfs_t size to be 120 bytes, got %d", sz)
	}
}

func TestFlockConstants(t *testing.T) {
	if LOCK_SH != 1 || LOCK_EX != 2 || LOCK_NB != 4 || LOCK_UN != 8 {
		t.Errorf("invalid flock constants")
	}
}
