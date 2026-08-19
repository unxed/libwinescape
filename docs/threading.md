# Go Runtime & Threading Considerations

## The Challenge

Standard Go `syscall` invocations are wrapped by runtime hooks (`entersyscall` / `exitsyscall`) which release the logical processor (`P`) so other goroutines can continue execution while an OS thread is blocked.

Because `libwinescape` invokes raw assembly `SYSCALL` instructions without cgo and without private runtime hooks, the Go scheduler is unaware that the OS thread has entered the kernel. If a blocking syscall (such as reading from a pipe or slow socket) is issued directly on an arbitrary goroutine, the OS thread and its associated `P` remain blocked.

Key implications:
1. **GOMAXPROCS Sizing:** For applications doing concurrent raw blocking I/O, `GOMAXPROCS` should be configured to ensure sufficient spare `P` processors remain available for other runnable goroutines.
2. **Stop-The-World Awareness:** Avoid triggering global `stopTheWorld` operations (e.g. `runtime.ReadMemStats` or heavy allocations that force STW GC) while waiting on inter-goroutine unblocking rendezvous across raw syscalls.

## The Solution: `winescape/gort`

For non-blocking filesystem I/O on local files, direct calls have minimal latency.

For operations that may block or when strict scheduler isolation is required, `libwinescape` provides the `gort` subpackage:
- A worker pool of goroutines bound to OS threads via `runtime.LockOSThread()`.
- Explicit bounded concurrency.
- Pure Go implementation without cgo.

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/unxed/libwinescape/go"
	"github.com/unxed/libwinescape/go/gort"
)

func main() {
	// Create a worker pool with 4 OS-thread-locked workers
	pool := gort.NewPool(gort.WithWorkers(4))
	defer pool.Close()

	// Execute a filesystem call safely inside the worker pool
	fd, err := gort.RunInPool(pool, func() (int, error) {
		return winescape.Open("/etc/hosts", winescape.O_RDONLY, 0)
	})
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer winescape.Close(fd)

	buf := make([]byte, 128)
	n, err := gort.RunInPool(pool, func() (int, error) {
		return winescape.Read(fd, buf)
	})
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	fmt.Printf("Read %d bytes: %s\n", n, string(buf[:n]))
}
```
