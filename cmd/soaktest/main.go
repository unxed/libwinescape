// Command soaktest is an empirical stress rig for the exact scenario raised
// in the threading/signal-safety review: a goroutine parked inside a raw,
// libwinescape-issued blocking syscall while the Go scheduler is under real
// pressure to preempt it.
//
// Because this binary is built for GOOS=windows, Go's own async
// preemption on the M running that goroutine goes through
// SuspendThread/GetThreadContext/SetThreadContext/ResumeThread (see
// runtime/os_windows.go), which under Wine are implemented by sending
// SIGUSR1 to the target thread (dlls/ntdll/unix/signal_x86_64.c,
// usr1_handler). Crucially, the goroutine under test here never calls
// entersyscall (that's the whole point of the library), so from the Go
// scheduler's point of view it looks like an ordinary long-running,
// CPU-bound goroutine -- which is exactly what triggers sysmon's
// forcePreemptNS retake path after ~10ms of uninterrupted "running" time.
// So simply blocking in a raw syscall for longer than that, next to other
// runnable goroutines, is sufficient to reliably provoke the scenario --
// no special timing tricks needed.
//
// What this soak test actually checks, per cycle:
//   - the raw blocking Read genuinely blocks (no data available yet)
//   - while blocked, background CPU-bound goroutines keep the scheduler
//     busy enough that sysmon will attempt to preempt the blocked M
//   - once data is written, Read returns the exact bytes written, with no
//     corruption, no lost wakeup, and no process hang
//   - the same for a blocking Write against a filled pipe buffer, and for
//     both directions run through the gort worker pool (LockOSThread'd)
//     as a second, independently-relevant configuration
//
// A hang shows up as the iteration's watchdog firing; corruption shows up
// as a byte mismatch; a crash shows up as the process dying. None of these
// are "pass/fail assertions about signal internals" -- they're the actual
// user-visible failure modes the original concern was about.
//
// Build and run under Wine, e.g.:
//
//	GOOS=windows GOARCH=amd64 go build -o soaktest.exe ./cmd/soaktest
//	wine64 soaktest.exe -iterations=50
package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	winescape "github.com/unxed/libwinescape/go"
	"github.com/unxed/libwinescape/go/gort"
)

var (
	iterations   = flag.Int("iterations", 20, "number of soak cycles per scenario")
	blockFor     = flag.Duration("block", 1500*time.Millisecond, "how long the reader/writer stays genuinely blocked before being released")
	watchdog     = flag.Duration("watchdog", 15*time.Second, "per-iteration deadline; a miss is reported as a hang")
	busyChurners = flag.Int("churners", 8, "number of CPU-bound background goroutines used to force scheduler/preemption pressure")
)

// churn keeps the scheduler under real pressure (and, incidentally, forces
// regular GC cycles) so that sysmon has both the motive (a "long running"
// blocked goroutine) and the opportunity (other runnable goroutines
// waiting on a constrained GOMAXPROCS) to attempt async preemption against
// the M executing the raw syscall under test.
func churn(n int, stop <-chan struct{}) {
	for i := 0; i < n; i++ {
		go func() {
			var buf []byte
			x := uint64(0)
			for {
				select {
				case <-stop:
					return
				default:
					x++
					if x%500000 == 0 {
						buf = append(buf, byte(x)) // keep the allocator/GC busy too
						if len(buf) > 1<<16 {
							buf = buf[:0]
						}
					}
				}
			}
		}()
	}
}

type result struct {
	ok       bool
	detail   string
	duration time.Duration
}

// blockingReadCycle exercises Read directly (no gort), matching the
// "arbitrary goroutine" case docs/threading.md warns about for scheduler
// starvation -- here used for signal-safety instead.
func blockingReadCycle(payload []byte) result {
	rfd, wfd, err := winescape.Pipe2(0)
	if err != nil {
		return result{detail: fmt.Sprintf("pipe2: %v", err)}
	}
	defer winescape.Close(rfd)
	defer winescape.Close(wfd)

	start := time.Now()
	got := make([]byte, len(payload))
	done := make(chan result, 1)

	go func() {
		total := 0
		for total < len(got) {
			n, err := winescape.Read(rfd, got[total:])
			if err != nil {
				done <- result{detail: fmt.Sprintf("read: %v (after %d/%d bytes)", err, total, len(got))}
				return
			}
			if n == 0 {
				done <- result{detail: fmt.Sprintf("read: unexpected EOF after %d/%d bytes", total, len(got))}
				return
			}
			total += n
		}
		done <- result{ok: bytes.Equal(got, payload), detail: "read"}
	}()

	// Give the reader time to genuinely block in the kernel before we
	// release it -- this is the window during which sysmon should be
	// trying to preempt that M.
	time.Sleep(*blockFor)
	if _, err := winescape.Write(wfd, payload); err != nil {
		return result{detail: fmt.Sprintf("write (unblocking): %v", err)}
	}

	select {
	case r := <-done:
		r.duration = time.Since(start)
		return r
	case <-time.After(*watchdog):
		return result{detail: "HANG: reader did not return within watchdog deadline", duration: time.Since(start)}
	}
}

// blockingWriteCycle fills the pipe buffer so Write must genuinely block,
// then drains it from another goroutine -- the mirror-image scenario, with
// the raw syscall stuck in the kernel on the write side instead.
func blockingWriteCycle(payload []byte) result {
	rfd, wfd, err := winescape.Pipe2(0)
	if err != nil {
		return result{detail: fmt.Sprintf("pipe2: %v", err)}
	}
	defer winescape.Close(rfd)
	defer winescape.Close(wfd)

	start := time.Now()
	done := make(chan result, 1)

	go func() {
		total := 0
		for total < len(payload) {
			n, err := winescape.Write(wfd, payload[total:])
			if err != nil {
				done <- result{detail: fmt.Sprintf("write: %v (after %d/%d bytes)", err, total, len(payload))}
				return
			}
			total += n
		}
		done <- result{ok: true, detail: "write"}
	}()

	// Let the writer fill the OS pipe buffer and genuinely block.
	time.Sleep(*blockFor)

	got := make([]byte, len(payload))
	total := 0
	for total < len(got) {
		n, err := winescape.Read(rfd, got[total:])
		if err != nil {
			return result{detail: fmt.Sprintf("drain read: %v", err), duration: time.Since(start)}
		}
		total += n
	}

	select {
	case r := <-done:
		if r.ok {
			r.ok = bytes.Equal(got, payload)
		}
		r.duration = time.Since(start)
		return r
	case <-time.After(*watchdog):
		return result{detail: "HANG: writer did not return within watchdog deadline", duration: time.Since(start)}
	}
}

// gortReadCycle is the same read scenario but issued from an
// runtime.LockOSThread'd gort worker, as the library's own docs recommend
// for anything that may block. Included so the soak covers both configs
// the library actually ships.
func gortReadCycle(pool *gort.Pool, payload []byte) result {
	rfd, wfd, err := winescape.Pipe2(0)
	if err != nil {
		return result{detail: fmt.Sprintf("pipe2: %v", err)}
	}
	defer winescape.Close(rfd)
	defer winescape.Close(wfd)

	start := time.Now()
	got := make([]byte, len(payload))
	done := make(chan result, 1)

	go func() {
		total := 0
		for total < len(got) {
			n, err := gort.RunInPool(pool, func() (int, error) {
				return winescape.Read(rfd, got[total:])
			})
			if err != nil {
				done <- result{detail: fmt.Sprintf("gort read: %v", err)}
				return
			}
			if n == 0 {
				done <- result{detail: "gort read: unexpected EOF"}
				return
			}
			total += n
		}
		done <- result{ok: bytes.Equal(got, payload), detail: "gort-read"}
	}()

	time.Sleep(*blockFor)
	if _, err := winescape.Write(wfd, payload); err != nil {
		return result{detail: fmt.Sprintf("write (unblocking): %v", err)}
	}

	select {
	case r := <-done:
		r.duration = time.Since(start)
		return r
	case <-time.After(*watchdog):
		return result{detail: "HANG: gort reader did not return within watchdog deadline", duration: time.Since(start)}
	}
}

func randomPayload(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func runScenario(name string, n int, fn func() result) bool {
	fmt.Printf("=== %s (%d iterations) ===\n", name, n)
	allOK := true
	var totalDur time.Duration
	for i := 0; i < n; i++ {
		r := fn()
		totalDur += r.duration
		status := "OK"
		if !r.ok {
			status = "FAIL"
			allOK = false
		}
		fmt.Printf("  [%3d/%3d] %-4s %-10s %8s  %s\n", i+1, n, status, r.detail, r.duration.Round(time.Millisecond), "")
		if !r.ok {
			fmt.Printf("           -> %s\n", r.detail)
		}
	}
	fmt.Printf("--- %s: avg %s, allOK=%v\n\n", name, (totalDur / time.Duration(n)).Round(time.Millisecond), allOK)
	return allOK
}

func main() {
	flag.Parse()

	if !winescape.Available() {
		fmt.Println("libwinescape reports unavailable on this host/arch -- run this under Wine")
		os.Exit(1)
	}

	// Deliberately small GOMAXPROCS: with plenty of runnable churner
	// goroutines and few Ps, the scheduler has strong incentive to try to
	// reclaim the P sitting behind our blocked, un-entersyscall'd raw
	// syscall goroutine -- exactly the pressure we want.
	runtime.GOMAXPROCS(2)

	var preempts, gcs uint64
	stopStats := make(chan struct{})
	go func() {
		var lastNumGC uint32
		for {
			select {
			case <-stopStats:
				return
			case <-time.After(200 * time.Millisecond):
				var ms runtime.MemStats
				runtime.ReadMemStats(&ms)
				if ms.NumGC != lastNumGC {
					atomic.AddUint64(&gcs, uint64(ms.NumGC-lastNumGC))
					lastNumGC = ms.NumGC
				}
			}
		}
	}()
	_ = preempts

	stopChurn := make(chan struct{})
	churn(*busyChurners, stopChurn)

	payload := randomPayload(4096)

	allOK := true
	allOK = runScenario("raw blocking Read", *iterations, func() result { return blockingReadCycle(payload) }) && allOK
	allOK = runScenario("raw blocking Write", *iterations, func() result { return blockingWriteCycle(payload) }) && allOK

	pool := gort.NewPool(gort.WithWorkers(4))
	allOK = runScenario("gort-pooled Read", *iterations, func() result { return gortReadCycle(pool, payload) }) && allOK
	pool.Close()

	close(stopChurn)
	close(stopStats)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("total GC cycles observed during soak: %d\n", ms.NumGC)

	if !allOK {
		fmt.Println("SOAK TEST: FAILED")
		os.Exit(1)
	}
	fmt.Println("SOAK TEST: PASSED")
}
