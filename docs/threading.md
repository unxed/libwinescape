# Go Runtime & Threading Considerations

## The Challenge

Standard Go `syscall` invocations are wrapped by runtime hooks (`entersyscall` / `exitsyscall`) which release the logical processor (`P`) so other goroutines can continue execution while an OS thread is blocked.

Because `libwinescape` invokes raw assembly `SYSCALL` instructions without cgo and without private runtime hooks, the Go scheduler is unaware that the OS thread has entered the kernel. If a blocking syscall (such as reading from a pipe or slow socket) is issued directly on an arbitrary goroutine, the OS thread and its associated `P` remain blocked.

## The Solution: `winescape/gort`

For non-blocking filesystem I/O on local files, direct calls have minimal latency.

For operations that may block or when strict scheduler isolation is required, `libwinescape` provides the `gort` subpackage:
- A worker pool of goroutines bound to OS threads via `runtime.LockOSThread()`.
- Explicit bounded concurrency.
- Pure Go implementation without cgo.
