#include "textflag.h"

// func syscall6_raw(nr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr)
TEXT ·syscall6_raw(SB), NOSPLIT, $0-80
	MOVD nr+0(FP), R8
	MOVD a1+8(FP), R0
	MOVD a2+16(FP), R1
	MOVD a3+24(FP), R2
	MOVD a4+32(FP), R3
	MOVD a5+40(FP), R4
	MOVD a6+48(FP), R5
	SVC $0
	MOVD $-4095, R7
	CMP R7, R0
	BHS is_err
	MOVD R0, r1+56(FP)
	MOVD $0, r2+64(FP)
	MOVD $0, err+72(FP)
	RET
is_err:
	NEG R0, R0
	MOVD $-1, R7
	MOVD R7, r1+56(FP)
	MOVD $0, r2+64(FP)
	MOVD R0, err+72(FP)
	RET
