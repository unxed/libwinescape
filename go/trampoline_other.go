//go:build !windows || (!amd64 && !arm64)

package winescape

import "syscall"

func syscall6_raw(nr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	return 0, 0, uintptr(syscall.ENOSYS)
}
