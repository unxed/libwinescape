# Risk review: FarManager/FarManager#1123 critique — 2026-08-19

Source: [FarGroup/FarManager#1123](https://github.com/FarGroup/FarManager/issues/1123),
comment from `johnd0e` relaying a friend's critique of the raw-syscall-under-Wine
approach. Reproduced here point by point, checked against what this repository
actually contains today — not against what the technique could theoretically
get wrong in the abstract.

## 1. "Wine doesn't guarantee register/ABI preservation, inserts trampolines/thunks"

**Partially valid as a general warning; not valid against what this library
actually does.** The critique is correct about *arbitrary* Win32→native
transitions — Wine's compatibility shims, exception machinery, and thunks are
real and do rewrite context in the general case. But `libwinescape` never asks
Wine to participate in that transition at all: it issues a bare `SYSCALL`
instruction from ordinary `.text` that Wine has no reason to recognize as
anything other than normal user-mode code. `docs/STATUS-signal-safety-review.md`
finding 3 traces this precisely through Wine's own source
(`dlls/ntdll/unix/signal_x86_64.c`, `is_inside_syscall()`): an interrupted raw
trampoline takes the *generic* suspend/resume path, not a Wine-specific one,
because the PC genuinely isn't inside anything Wine instrumented.
**Verdict: addressed, with a source-level citation, not just an assertion.**

## 2. "Go runtime is aggressive: goroutine migration, signals (SIGURG/SIGPROF), stack growth"

**The single most valid point in the critique, and it was a real, unclosed gap
at the time it was raised.** `docs/threading.md` already documented the
scheduler-starvation half (P/M staying blocked because raw syscalls bypass
`entersyscall`). It did **not** originally cover the signal-preemption half —
whether an async-preemption signal landing *mid-syscall* could corrupt
register/PC state. That gap is what `docs/STATUS-signal-safety-review.md` was
written to close, and it closes it three ways, not one:
- **By construction:** `runtime.isAsyncSafePoint` refuses to preempt inside a
  hand-written asm function with no `FUNCDATA_LocalsPointerMaps`
  (`syscall6_raw` qualifies), so Go's own preemption logic can't land mid-trap
  regardless of OS.
- **By platform fact, checked against Go source, not assumed:** this library
  targets `GOOS=windows`, and the Windows-target Go runtime doesn't use POSIX
  signals for async preemption at all (`SuspendThread`/`GetThreadContext`, not
  `SIGURG`). `cmd/sigprobe` confirmed this empirically under real Wine —
  `SIGURG`/`SIGPROF` stay `SIG_DFL` at every checkpoint.
- **Empirically, under load:** `cmd/soaktest` (raw blocking Read/Write on
  pipes, gort-pooled Read under deliberate CPU-bound scheduler pressure) ran
  hundreds of iterations under real Wine with zero hangs and zero corruption.
  This is exactly the "run it and see what happens when a signal lands
  mid-syscall" experiment the critique itself said was missing.

**Verdict: was genuinely open when raised; is now closed, empirically, not
just argued. The one real bug the review surfaced along the way — inconsistent
EINTR retry across wrappers — is already fixed (`docs/STATUS-signal-safety-review.md`
§5).**

## 3. "Syscall numbers/structures differ across kernels and architectures"

**Correct as a general fact, not a critique of this design.** This isn't a
flaw the raw-syscall approach introduces — it's the ordinary cost of
supporting more than one kernel ABI, and it's exactly what `spec/table.go`
plus `cmd/gen-numbers` exist to contain: numbers come from kernel headers,
never invented, one entry per (syscall, OS, arch), regenerated mechanically so
the Go and C sides can't drift apart. The FreeBSD carry-flag error convention
— a genuinely different ABI from Linux's negative-`errno`-in-`rax`, not just a
different number table — is handled as its own code path, not assumed to
match Linux's.

**One real, still-open gap this point does correctly gesture at:** NetBSD and
DragonFly are marked "presumed BSD ABI, pending live hardware verification" in
the platform table (`README.md`). That's an honest, already-flagged unknown,
not a claim of certainty — see "Open items" below.

## 4. "Wine could change its internals in any version without warning"

**Correct, and this is a real, standing risk — not something to argue away.**
Nothing about this library's design makes Wine's `ntdll`/Unix-call boundary a
documented, versioned contract; it happens to be stable today because raw
syscalls sit *below* everything Wine controls, but a hypothetical future Wine
that started intercepting raw `SYSCALL` traps generically (e.g. for
sandboxing) would break this without deprecation notice.

**Mitigated, as of 2026-08-19, by an automatic startup self-test** rather than
only by the manual `cmd/soaktest`/`cmd/sigprobe` tools a human has to
remember to run. `Available()` now round-trips real data through the exact
sequence its callers depend on — create/write/close/reopen/read/compare/
unlink a real temp file, plus a deliberate open-of-missing-file check that
confirms errno translation itself is correct, not just "some syscall
returned success" — once, memoized, on first call. See `go/selftest_windows.go`
for exactly what this does and doesn't catch: it catches the trampoline not
reaching the kernel, wrong syscall numbers for the operations it exercises,
and broken errno translation; it does **not** catch anything that only shows
up under concurrency/signal pressure (that stays `cmd/soaktest`'s job, and is
too slow to run at every startup) or bugs in syscalls the test doesn't happen
to exercise. If a future Wine breaks the raw-syscall path, `Available()`
reports `false` and every caller falls back to the safe Win32 path
automatically, instead of a human discovering corruption after the fact.
`SelfTestError()` exposes the specific failure for logging.

What remains, and is accepted as a standing, monitored risk with no one-time
technical fix: keeping `cmd/soaktest`/`cmd/sigprobe` runnable so a Wine
upgrade can still be checked deliberately under load, and keeping the
platform-support table honest about what's verified versus presumed (see
item 3).

## 5. "It works because Wine ultimately runs a normal ELF process, and if you reach a genuine Linux thread without touching Win32 structures, the kernel doesn't know the process is 'Windows'"

This is the critique's own explanation of *why the technique works at all*,
not an objection — and it's an accurate one-sentence description of this
library's entire premise. Worth keeping on record as confirmation that the
critique's author understood the mechanism correctly even while raising
concerns about its edges.

## Open items (not closed by this review, recorded so they aren't lost)

- NetBSD/DragonFly ABI: presumed, not verified on real hardware (README
  platform table already says so).
- BSD-family ARM64: not probed at all.
- Wine-version monitoring is a process commitment (re-run the soak test
  against new Wine releases), not a one-time technical fix — there is no
  patch that closes this permanently.

## Conclusion posted back to the issue

The one-paragraph response already given in the GitHub thread — reporting the
soak-test results and noting that Go-runtime nuances don't apply to native
C/C++ callers like Far3 at all — is accurate and doesn't need amending. This
document exists so the reasoning behind that conclusion is checkable from the
repository itself, not only from a GitHub comment.
