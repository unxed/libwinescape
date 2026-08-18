package winescape

import (
	"testing"
	"unsafe"
)

func TestWinsize_Size(t *testing.T) {
	var ws Winsize
	sz := unsafe.Sizeof(ws)
	if sz != 8 {
		t.Errorf("expected Winsize size 8 bytes, got %d", sz)
	}
}

func TestClockConstants(t *testing.T) {
	if CLOCK_REALTIME != 0 || CLOCK_MONOTONIC != 1 {
		t.Errorf("invalid clock constants: REALTIME=%d, MONOTONIC=%d", CLOCK_REALTIME, CLOCK_MONOTONIC)
	}
	if SIGWINCH != 28 {
		t.Errorf("expected SIGWINCH = 28, got %d", SIGWINCH)
	}
}
