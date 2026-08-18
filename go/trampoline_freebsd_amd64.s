//go:build windows && amd64

#include "textflag.h"

// func syscall6_raw(nr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr)
TEXT ·syscall6_raw_freebsd(SB), NOSPLIT, $0-80
	MOVQ nr+0(FP), AX
	MOVQ a1+8(FP), DI
	MOVQ a2+16(FP), SI
	MOVQ a3+24(FP), DX
	MOVQ a4+32(FP), R10
	MOVQ a5+40(FP), R8
	MOVQ a6+48(FP), R9
	SYSCALL
	JC is_err
	MOVQ AX, r1+56(FP)
	MOVQ $0, r2+64(FP)
	MOVQ $0, err+72(FP)
	RET
is_err:
	MOVQ $-1, r1+56(FP)
	MOVQ $0, r2+64(FP)
	MOVQ AX, err+72(FP)
	RET
