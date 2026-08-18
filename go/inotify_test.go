package winescape

import (
	"encoding/binary"
	"testing"
)

func TestParseInotifyEvents_Synthetic(t *testing.T) {
	// Build a synthetic 32-byte inotify buffer:
	// wd=1, mask=IN_CREATE (0x100), cookie=0, len=16, name="newfile.txt\0..."
	buf := make([]byte, 32)
	binary.LittleEndian.PutUint32(buf[0:4], 1)
	binary.LittleEndian.PutUint32(buf[4:8], IN_CREATE)
	binary.LittleEndian.PutUint32(buf[8:12], 0)
	binary.LittleEndian.PutUint32(buf[12:16], 16)
	copy(buf[16:32], "newfile.txt\x00")

	events, err := ParseInotifyEvents(buf)
	if err != nil {
		t.Fatalf("ParseInotifyEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Wd != 1 || events[0].Mask != IN_CREATE || events[0].Name != "newfile.txt" {
		t.Errorf("event mismatch: %+v", events[0])
	}
}
