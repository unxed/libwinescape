package winescape

import (
	"testing"
	"unsafe"
)

func TestTermios_Size(t *testing.T) {
	var term Termios
	sz := unsafe.Sizeof(term)
	if sz != 60 {
		t.Errorf("expected Termios size 60 bytes, got %d", sz)
	}
}

func TestTermiosConstants(t *testing.T) {
	if TCGETS != 0x5401 || TCSETS != 0x5402 {
		t.Errorf("invalid TCGETS/TCSETS constants")
	}
	if CS8 != 0000060 || ICANON != 0000002 {
		t.Errorf("invalid termios flag constants")
	}
}
