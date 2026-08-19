//go:build windows

package winescape

import (
	"bytes"
	"errors"
	"fmt"
	"syscall"
)

// selfTest exercises, once, the actual sequence of raw syscalls this
// library's real callers depend on -- not just "does a trap reach the
// kernel at all" (cmd/sigprobe already answers that, once, by hand, for a
// human to read) but "does round-tripping real data through it, right now,
// on this Wine version, produce correct results". This is the automatic,
// always-on counterpart to that manual probe: it runs on the first call to
// Available() (memoized, like everything else in this file), costs a
// handful of syscalls (microseconds), and its result is what
// docs/RISK-REVIEW-farmanager-1123.md item 4 points to as the mitigation
// for "Wine could change its raw-syscall passthrough behavior in a future
// version without warning" -- a broken future Wine fails this and
// Available() reports false, so callers fall back to the safe Win32 path
// automatically instead of a human noticing corruption after the fact.
//
// What this catches: the trampoline not reaching the kernel at all, wrong
// syscall numbers for open/write/read/close/unlink specifically, and
// errno translation being wrong (via the deliberate ENOENT check).
//
// What this does NOT catch, and is not a substitute for: anything that
// only manifests under concurrency or signal pressure (that's
// cmd/soaktest's job, and it's too slow to run at every startup),
// correctness of syscalls this test doesn't happen to exercise, or
// anything the test's small write/read size wouldn't trigger (e.g. a
// hypothetical future bug specific to very large transfers).
func selfTest() error {
	const probeDir = "/tmp"
	const payload = "libwinescape-selftest"

	f, err := CreateTemp(probeDir, "wsprobe-*")
	if err != nil {
		return fmt.Errorf("selfTest: CreateTemp: %w", err)
	}
	path := f.Name()
	defer Unlink(path) // best-effort; the explicit Unlink below is the real check

	if _, err := f.Write([]byte(payload)); err != nil {
		f.Close()
		return fmt.Errorf("selfTest: write: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("selfTest: close after write: %w", err)
	}

	rf, err := Open(path, O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("selfTest: reopen for read: %w", err)
	}
	buf := make([]byte, len(payload)+8)
	n, err := Read(rf, buf)
	closeErr := Close(rf)
	if err != nil {
		return fmt.Errorf("selfTest: read: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("selfTest: close after read: %w", closeErr)
	}
	if !bytes.Equal(buf[:n], []byte(payload)) {
		return fmt.Errorf("selfTest: round-trip mismatch: wrote %q, read back %q", payload, buf[:n])
	}

	if err := Unlink(path); err != nil {
		return fmt.Errorf("selfTest: unlink: %w", err)
	}

	// Deliberate-failure path: confirms errno translation actually produces
	// the specific error real callers branch on (os.IsNotExist-shaped
	// checks throughout f4's vfs/hostfs), not just "some error came back".
	if _, err := Open(path, O_RDONLY, 0); err == nil {
		return fmt.Errorf("selfTest: open of just-deleted file unexpectedly succeeded")
	} else if !errors.Is(err, syscall.ENOENT) {
		return fmt.Errorf("selfTest: open of missing file returned %v, want ENOENT-shaped error", err)
	}

	pid, err := Getpid()
	if err != nil {
		return fmt.Errorf("selfTest: getpid: %w", err)
	}
	if pid <= 0 || pid >= 4194304 {
		return fmt.Errorf("selfTest: getpid returned implausible value %d", pid)
	}

	return nil
}
