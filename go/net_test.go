package winescape

import (
	"testing"
	"unsafe"
)

func TestRawSockaddrUn_Size(t *testing.T) {
	var sa RawSockaddrUn
	sz := unsafe.Sizeof(sa)
	if sz != 110 {
		t.Errorf("expected RawSockaddrUn size 110 bytes, got %d", sz)
	}
}
