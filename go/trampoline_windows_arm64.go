//go:build windows && arm64

package winescape

func syscall6_raw(nr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr)
