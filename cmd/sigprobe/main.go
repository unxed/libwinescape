// Command sigprobe answers one narrow, empirical question: does Wine ever
// install or restore its own sigaction for SIGURG (Go's async-preemption
// signal) or SIGPROF (Go's CPU profiler signal) after the Go runtime's own
// initsig() has already installed its handlers?
//
// It is not a general-purpose test tool: it prints raw kernel_sigaction
// snapshots for both signals at a few points in the program's life (at
// startup, after doing real raw-syscall filesystem I/O through
// libwinescape, under heavy goroutine/scheduler pressure, and after that
// pressure stops) and leaves the comparison to whoever reads the output.
// If the handler address ever changes after Go's own startup, something
// downstream of the Go runtime is touching these signals and the finding
// needs to be investigated properly; if it never changes, this specific
// concern is empirically closed for the Wine/kernel combination it was run
// against.
//
// Build and run under Wine, e.g.:
//
//	GOOS=windows GOARCH=amd64 go build -o sigprobe.exe ./cmd/sigprobe
//	wine64 sigprobe.exe
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	winescape "github.com/unxed/libwinescape/go"
)

// Linux/x86-64 syscall number and signal numbers. Not routed through
// spec/table.go because this is a standalone diagnostic, not a supported
// library call.
const (
	sysRtSigaction = 13
	sigURG         = 23
	sigPROF        = 27
)

// kernelSigaction matches the Linux kernel's struct kernel_sigaction ABI
// (NOT glibc's struct sigaction, which has a different, libc-specific
// layout). sa_mask is a single 8-byte sigset_t as the kernel expects when
// sigsetsize == 8.
type kernelSigaction struct {
	Handler  uintptr
	Flags    uint64
	Restorer uintptr
	Mask     uint64
}

func getSigaction(sig int) (kernelSigaction, error) {
	var out kernelSigaction
	// rt_sigaction(sig, NULL, &out, sizeof(sigset_t)) -- query only, install nothing.
	_, _, err := winescape.Syscall6(sysRtSigaction, uintptr(sig), 0, uintptr(unsafe.Pointer(&out)), 8, 0, 0)
	return out, err
}

func report(label string) {
	urg, errU := getSigaction(sigURG)
	prof, errP := getSigaction(sigPROF)
	fmt.Printf("[%-14s] SIGURG  handler=0x%016x flags=0x%x restorer=0x%016x err=%v\n",
		label, urg.Handler, urg.Flags, urg.Restorer, errU)
	fmt.Printf("[%-14s] SIGPROF handler=0x%016x flags=0x%x restorer=0x%016x err=%v\n",
		label, prof.Handler, prof.Flags, prof.Restorer, errP)
}

func busyGoroutines(n int, stop <-chan struct{}) {
	for i := 0; i < n; i++ {
		go func() {
			x := 0
			for {
				select {
				case <-stop:
					return
				default:
					x++
				}
			}
		}()
	}
}

func main() {
	if !winescape.Available() {
		fmt.Println("libwinescape reports unavailable on this host/arch -- run this under Wine on Linux/FreeBSD amd64 or arm64")
		os.Exit(1)
	}

	report("startup")

	// Exercise the exact code path the risk is about: a real raw syscall
	// (not just the sigaction probe itself) issued through the trampoline.
	fd, err := winescape.Open("/etc/hostname", winescape.O_RDONLY, 0)
	if err != nil {
		fmt.Println("note: could not open /etc/hostname for the I/O exercise:", err)
	} else {
		buf := make([]byte, 64)
		winescape.Read(fd, buf)
		winescape.Close(fd)
	}

	report("after-raw-io")

	// Force real scheduler pressure so the runtime actually needs async
	// preemption (SIGURG) to make progress: many CPU-bound goroutines on a
	// deliberately small GOMAXPROCS.
	runtime.GOMAXPROCS(2)
	stop := make(chan struct{})
	busyGoroutines(8, stop)

	// Force SIGPROF traffic too (100 Hz while active).
	profFile, ferr := os.CreateTemp("", "sigprobe-*.pprof")
	if ferr == nil {
		pprof.StartCPUProfile(profFile)
	}

	time.Sleep(2 * time.Second)
	report("under-load")

	if ferr == nil {
		pprof.StopCPUProfile()
		profFile.Close()
		os.Remove(profFile.Name())
	}
	close(stop)
	time.Sleep(200 * time.Millisecond)

	report("after-load")

	fmt.Println("done")
}
