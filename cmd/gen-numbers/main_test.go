package main

import (
	"strings"
	"testing"

	"github.com/unxed/libwinescape/spec"
)

func TestGenerateGoFile_LinuxAMD64(t *testing.T) {
	out := generateGoFile(spec.OSLinux, spec.ArchAMD64, "linux && amd64")
	if !strings.Contains(out, "package winescape") {
		t.Errorf("missing package statement in generated Go file")
	}
	if !strings.Contains(out, "sysRead uintptr = 0") {
		t.Errorf("expected sysRead = 0 for linux/amd64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysGetpid uintptr = 39") {
		t.Errorf("expected sysGetpid = 39 for linux/amd64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysGetdents64 uintptr = 217") {
		t.Errorf("expected sysGetdents64 = 217 for linux/amd64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysGetdirentries uintptr = 0") {
		t.Errorf("expected sysGetdirentries = 0 for linux/amd64 fallback, got output:\n%s", out)
	}
}

func TestGenerateGoFile_LinuxARM64(t *testing.T) {
	out := generateGoFile(spec.OSLinux, spec.ArchARM64, "linux && arm64")
	if !strings.Contains(out, "sysRead uintptr = 63") {
		t.Errorf("expected sysRead = 63 for linux/arm64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysGetpid uintptr = 172") {
		t.Errorf("expected sysGetpid = 172 for linux/arm64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysOpenat uintptr = 56") {
		t.Errorf("expected sysOpenat = 56 for linux/arm64, got output:\n%s", out)
	}
}

func TestGenerateGoFile_FreeBSDAMD64(t *testing.T) {
	out := generateGoFile(spec.OSFreeBSD, spec.ArchAMD64, "freebsd && amd64")
	if !strings.Contains(out, "sysRead uintptr = 3") {
		t.Errorf("expected sysRead = 3 for freebsd/amd64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysGetpid uintptr = 20") {
		t.Errorf("expected sysGetpid = 20 for freebsd/amd64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysGetdirentries uintptr = 554") {
		t.Errorf("expected sysGetdirentries = 554 for freebsd/amd64, got output:\n%s", out)
	}
	if !strings.Contains(out, "sysGetdents64 uintptr = 0") {
		t.Errorf("expected sysGetdents64 = 0 for freebsd/amd64 fallback, got output:\n%s", out)
	}
}

func TestGenerateCHeader(t *testing.T) {
	out := generateCHeader()
	if !strings.Contains(out, "#ifndef WINESCAPE_NUMBERS_H") {
		t.Errorf("missing header guard in generated C header")
	}
	if !strings.Contains(out, "#define WS_SYS_GETPID 39") {
		t.Errorf("expected WS_SYS_GETPID 39 in Linux amd64 block, got:\n%s", out)
	}
	if !strings.Contains(out, "#define WS_SYS_GETPID 172") {
		t.Errorf("expected WS_SYS_GETPID 172 in Linux arm64 block, got:\n%s", out)
	}
	if !strings.Contains(out, "#define WS_SYS_GETPID 20") {
		t.Errorf("expected WS_SYS_GETPID 20 in FreeBSD amd64 block, got:\n%s", out)
	}
}
