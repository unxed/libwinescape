# Signal-safety review: goroutines vs. raw syscalls under Wine (status)

## Origin

A review raised this concern: Go's async preemption (SIGURG) and profiler
(SIGPROF) signals, plus Wine's own signal machinery, are an actively
mutating environment, and a single-threaded `getpid()` smoke test says
nothing about what happens if a signal lands while a goroutine is inside
`syscall6_raw`. `docs/threading.md` §"The Challenge" documents the
scheduler-starvation half of this; this note documents the
signal-safety half.

## Findings, in order

1. **Register/PC corruption mid-syscall is not possible by construction.**
   `runtime/preempt.go`'s `isAsyncSafePoint` refuses to async-preempt any
   PC inside a hand-written assembly function (`f.flag&abi.FuncFlagAsm != 0`,
   no `FUNCDATA_LocalsPointerMaps`). `syscall6_raw` is exactly such a
   function (`TEXT ·syscall6_raw(SB), NOSPLIT`). This holds regardless of
   OS, since `isAsyncSafePoint` is shared runtime code.

2. **SIGURG/SIGPROF are not the actual mechanism for this library.** This
   library targets `GOOS=windows` binaries. The Windows-target Go runtime
   does not use POSIX signals for async preemption at all --
   `runtime/os_windows.go`'s `preemptM` uses `SuspendThread` /
   `GetThreadContext` / `SetThreadContext` / `ResumeThread`. Confirmed
   empirically with `cmd/sigprobe`: SIGURG and SIGPROF stay `SIG_DFL`
   (handler `0x0`) at every checkpoint, including under deliberate
   scheduler pressure, across three runs under Wine 9.0 / Ubuntu 24.04.

3. **The real mechanism is Wine's own `SuspendThread` implementation.**
   Wine's `usr1_handler` (`dlls/ntdll/unix/signal_x86_64.c`) is what
   backs `SuspendThread`/`GetThreadContext` under Linux, via `SIGUSR1`. It
   branches on `is_inside_syscall(ucontext)`, which checks whether the
   interrupted PC is inside Wine's own `__wine_syscall_dispatcher`. Our
   raw trampoline never reaches that dispatcher (that's the point of the
   library), so Wine takes the generic branch: save full context, block in
   `wait_suspend()`, restore context, return -- i.e. it treats an
   interrupted raw syscall exactly like any other interrupted native
   instruction stream. If the raw `SYSCALL` was already dispatched into
   the kernel and blocked, `SIGUSR1` delivery yields the normal
   EINTR/ERESTART semantics to userspace.

4. **The one real, concrete gap found: inconsistent EINTR handling.**
   Before this review, only `Sleep`/`ClockNanosleep` retried on
   EINTR/ERESTART; every other blocking-capable wrapper
   (`Read`/`Write`/`Pread`/`Pwrite`/`Openat`/`Flock`/`Accept4`/`Wait4`)
   did not, and would surface a spurious error to callers under signal
   pressure. Fixed (already on `main` as of this note) via a `retryEINTR`
   helper, applied only where interruption-before-completion has no
   observable side effect. Deliberately NOT applied to `Close` (retrying
   `close()` on EINTR is a classic use-after-free-of-fd bug) or `Connect`
   (POSIX requires poll-for-writability + `SO_ERROR`, not a blind retry).

## What's verified vs. still open

- Verified by reading Go runtime source + Wine source: (1)-(3) above.
- Verified by a single empirical run under real Wine: `cmd/sigprobe`'s
  SIG_DFL finding (3 repeated runs, consistent).
- **Not yet verified**: a real soak run of `cmd/soaktest` (raw blocking
  Read/Write on a pipe, released after a delay, under CPU-bound scheduler
  pressure on `GOMAXPROCS(2)`, both direct and via the `gort` pool) for
  many iterations under Wine, checking for hangs, data corruption, or
  crashes. The binary builds and a short manual smoke run completed
  during this review, but a real multi-iteration soak was not run to
  completion (session ended). This is the next concrete step -- see
  `cmd/soaktest/main.go`'s own doc comment for how it's meant to be run,
  and the instructions below.

## How to run the soak test yourself

```sh
GOOS=windows GOARCH=amd64 go build -o soaktest.exe ./cmd/soaktest
wine64 soaktest.exe -iterations=200 -block=1500ms -watchdog=15s -churners=8
```

- `-iterations` -- cycles per scenario (raw Read, raw Write, gort-pooled
  Read). Start with 200+; this is cheap (each cycle is ~`block` seconds).
- `-block` -- how long the reader/writer is left genuinely blocked in the
  kernel before being released. Keep it well above sysmon's ~10ms
  `forcePreemptNS` window so a preemption attempt is essentially
  guaranteed per cycle; the default (1.5s) is already generous.
- `-watchdog` -- per-iteration deadline. If this fires, that's a hang --
  the process will print `HANG: ...` for that iteration and keep going,
  but treat any HANG line as a failing run.
- `-churners` -- number of CPU-bound background goroutines that keep the
  scheduler under pressure so sysmon actually has a reason to try
  preempting the blocked M. Leave at the default unless investigating.

**Reading the output:**
- Every line is `[i/N] STATUS detail duration`. `STATUS` is `OK` or
  `FAIL`. A `FAIL` line prints a second `-> ...` line explaining what
  went wrong (mismatch, HANG, syscall error).
- The process exits with status `1` and prints `SOAK TEST: FAILED` if any
  iteration in any of the three scenarios failed; otherwise it prints
  `SOAK TEST: PASSED` with exit code `0`. For a scripted check:
  `wine64 soaktest.exe -iterations=200 || echo "soak test failed"`.
- A `FAIL` with detail containing `HANG` means the raw syscall never
  returned after being unblocked -- the scenario this whole review is
  about. A `FAIL` with a byte-mismatch detail means data corruption. Both
  are exactly the failure modes worth caring about; everything else
  (a clean `SOAK TEST: PASSED` after a few hundred iterations across all
  three scenarios) is the empirical confirmation that finding 3 above
  actually holds in practice, not just on paper.

## Patches

- `0001-...`: EINTR/ERESTART retry fix (already merged to `main`).
- `0002-...`: `cmd/sigprobe` + `cmd/soaktest` (this patch).
