package winescape

import (
	"testing"
)

func TestErrno_Error(t *testing.T) {
	e0 := Errno(0)
	if e0.Error() != "errno 0 (success)" {
		t.Errorf("unexpected zero errno message: %s", e0.Error())
	}

	e2 := Errno(2) // ENOENT
	if e2.Error() == "" {
		t.Errorf("expected non-empty string for ENOENT (errno 2)")
	}
}

func TestSyscallConstants_NonZero(t *testing.T) {
	if sysWrite != 1 && sysWrite != 4 && sysWrite != 64 {
		t.Errorf("sysWrite number %d is not matching known platforms", sysWrite)
	}
	if sysClose != 3 && sysClose != 6 && sysClose != 57 {
		t.Errorf("sysClose number %d is not matching known platforms", sysClose)
	}
	if sysGetpid == 0 {
		t.Errorf("sysGetpid must not be 0")
	}
}
