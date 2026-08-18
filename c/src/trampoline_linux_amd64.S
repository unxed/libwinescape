#if defined(__linux__) && (defined(__x86_64__) || defined(_M_X64))

.text
.globl ws_syscall6
.def ws_syscall6; .scl 2; .type 32; .endef

/* intptr_t ws_syscall6(uintptr_t nr, uintptr_t a1, uintptr_t a2,
 *                      uintptr_t a3, uintptr_t a4, uintptr_t a5,
 *                      uintptr_t a6, int *err_out)
 * Under Windows x64 ABI:
 *   RCX = nr, RDX = a1, R8 = a2, R9 = a3
 *   [RSP+40] = a4, [RSP+48] = a5, [RSP+56] = a6, [RSP+64] = err_out
 */
ws_syscall6:
	pushq %rbx
	pushq %r12
	pushq %r13
	pushq %r14
	pushq %r15

	movq %rcx, %rax       /* syscall nr */
	movq %rdx, %rdi       /* a1 */
	movq %r8,  %rsi       /* a2 */
	movq %r9,  %rdx       /* a3 */
	movq 88(%rsp), %r10   /* a4: 40 + 5*8 push = 80 + 8 ret = 88 */
	movq 96(%rsp), %r8    /* a5 */
	movq 104(%rsp), %r9   /* a6 */

	syscall

	movq 112(%rsp), %r11  /* err_out pointer */
	cmpq $-4095, %rax
	jae 1f

	/* Success */
	testq %r11, %r11
	jz 2f
	movl $0, (%r11)
	jmp 2f

1:	/* Error */
	negq %rax
	testq %r11, %r11
	jz 3f
	movl %eax, (%r11)
3:	movq $-1, %rax

2:	popq %r15
	popq %r14
	popq %r13
	popq %r12
	popq %rbx
	ret

#endif
