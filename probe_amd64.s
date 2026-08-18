#include "textflag.h"

// func rawGetpid() uint64
TEXT ·rawGetpid(SB), NOSPLIT, $0-8
	MOVQ $39, AX
	SYSCALL
	MOVQ AX, ret+0(FP)
	RET
