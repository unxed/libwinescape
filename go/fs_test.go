package winescape

import (
	"encoding/binary"
	"syscall"
	"testing"
	"unsafe"
)

func TestBytePtrFromString(t *testing.T) {
	ptr, err := BytePtrFromString("hello/world")
	if err != nil {
		t.Fatalf("unexpected error for valid string: %v", err)
	}
	if ptr == nil {
		t.Fatal("expected non-nil pointer")
	}

	_, errNul := BytePtrFromString("hello\x00world")
	if errNul != syscall.EINVAL {
		t.Errorf("expected EINVAL for string with embedded NUL, got %v", errNul)
	}
}

func TestStat_t_Size(t *testing.T) {
	var st Stat_t
	size := unsafe.Sizeof(st)
	if size != 144 {
		t.Errorf("expected Stat_t size to be 144 bytes, got %d", size)
	}
}

func TestParseDirent64_Synthetic(t *testing.T) {
	// Build a synthetic 64-bit linux_dirent64 buffer:
	// Entry 1: ino=100, off=1, reclen=24, type=DT_DIR (4), name="."
	// Entry 2: ino=101, off=2, reclen=32, type=DT_REG (8), name="test.txt"
	buf := make([]byte, 56)

	// Entry 1 (offset 0..24)
	binary.LittleEndian.PutUint64(buf[0:8], 100)
	binary.LittleEndian.PutUint64(buf[8:16], 1)
	binary.LittleEndian.PutUint16(buf[16:18], 24)
	buf[18] = DT_DIR
	buf[19] = '.'
	buf[20] = 0

	// Entry 2 (offset 24..56)
	binary.LittleEndian.PutUint64(buf[24:32], 101)
	binary.LittleEndian.PutUint64(buf[32:40], 2)
	binary.LittleEndian.PutUint16(buf[40:42], 32)
	buf[42] = DT_REG
	copy(buf[43:52], "test.txt\x00")

	entries, err := ParseDirent64(buf)
	if err != nil {
		t.Fatalf("ParseDirent64 failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Name != "." || entries[0].Ino != 100 || entries[0].Type != DT_DIR {
		t.Errorf("unexpected entry 0: %+v", entries[0])
	}
	if entries[1].Name != "test.txt" || entries[1].Ino != 101 || entries[1].Type != DT_REG {
		t.Errorf("unexpected entry 1: %+v", entries[1])
	}
}
