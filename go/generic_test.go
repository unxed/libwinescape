package winescape

import (
	"testing"
)

func TestGenericSyscall_NumbersDefined(t *testing.T) {
	// Verify that key syscall numbers are present in the generic dispatch table.
	if sysGetpid == 0 {
		t.Errorf("sysGetpid must be non-zero on supported target platforms")
	}
	if sysRead == 0 && sysWrite == 0 && sysClose == 0 {
		t.Errorf("expected basic I/O syscall numbers to be initialized")
	}
}
