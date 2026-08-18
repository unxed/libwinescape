package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	winescape "github.com/unxed/libwinescape/go"
	"github.com/unxed/libwinescape/go/gort"
)

func main() {
	fmt.Println("=== libwinescape End-to-End Integration Test Suite ===")
	passed := 0
	failed := 0

	runTest := func(name string, fn func() error) {
		fmt.Printf("TEST %-50s ... ", name)
		if err := fn(); err != nil {
			fmt.Printf("FAIL (%v)\n", err)
			failed++
		} else {
			fmt.Println("PASS")
			passed++
		}
	}

	runTest("Host getpid syscall", func() error {
		pid, err := winescape.Getpid()
		if err != nil {
			return err
		}
		if pid <= 0 || pid > 4194304 {
			return fmt.Errorf("implausible pid: %d", pid)
		}
		fmt.Printf("[pid=%d] ", pid)
		return nil
	})

	testFile := fmt.Sprintf("/tmp/libwinescape_test_%d_%d.tmp", os.Getpid(), time.Now().UnixNano())
	payload := []byte("Hello from libwinescape raw POSIX syscalls under Wine!\n")

	runTest("Create and write file (/tmp)", func() error {
		fd, err := winescape.Open(testFile, winescape.O_RDWR|winescape.O_CREAT|winescape.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("open error: %w", err)
		}
		defer winescape.Close(fd)

		n, err := winescape.Write(fd, payload)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
		if n != len(payload) {
			return fmt.Errorf("short write: %d != %d", n, len(payload))
		}
		return nil
	})

	runTest("Seek and read back content", func() error {
		fd, err := winescape.Open(testFile, winescape.O_RDONLY, 0)
		if err != nil {
			return fmt.Errorf("open error: %w", err)
		}
		defer winescape.Close(fd)

		off, err := winescape.Seek(fd, 0, winescape.SEEK_SET)
		if err != nil {
			return fmt.Errorf("seek error: %w", err)
		}
		if off != 0 {
			return fmt.Errorf("unexpected seek offset: %d", off)
		}

		readBuf := make([]byte, len(payload)+16)
		n, err := winescape.Read(fd, readBuf)
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
		if !bytes.Equal(readBuf[:n], payload) {
			return fmt.Errorf("data mismatch: got %q, want %q", string(readBuf[:n]), string(payload))
		}
		return nil
	})

	runTest("Stat file verification", func() error {
		var st winescape.Stat_t
		err := winescape.Stat(testFile, &st)
		if err != nil {
			return fmt.Errorf("stat error: %w", err)
		}
		if st.Size != int64(len(payload)) {
			return fmt.Errorf("stat size mismatch: %d != %d", st.Size, len(payload))
		}
		return nil
	})

	runTest("Directory listing (getdents64 /tmp)", func() error {
		dirFd, err := winescape.Open("/tmp", winescape.O_RDONLY|winescape.O_DIRECTORY, 0)
		if err != nil {
			return fmt.Errorf("open /tmp error: %w", err)
		}
		defer winescape.Close(dirFd)

		buf := make([]byte, 8192)
		n, err := winescape.Getdents64(dirFd, buf)
		if err != nil {
			return fmt.Errorf("getdents64 error: %w", err)
		}
		if n <= 0 {
			return fmt.Errorf("getdents64 returned 0 bytes")
		}

		entries, err := winescape.ParseDirent64(buf[:n])
		if err != nil {
			return fmt.Errorf("ParseDirent64 error: %w", err)
		}

		targetBase := filepath.Base(testFile)
		found := false
		for _, e := range entries {
			if e.Name == targetBase {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("created file %q not found in directory listing of /tmp", targetBase)
		}
		return nil
	})

	runTest("Unlink temporary file", func() error {
		err := winescape.Unlink(testFile)
		if err != nil {
			return fmt.Errorf("unlink error: %w", err)
		}
		var st winescape.Stat_t
		if err := winescape.Stat(testFile, &st); err == nil {
			return fmt.Errorf("file still exists after unlink")
		}
		return nil
	})

	runTest("gort worker pool execution", func() error {
		pool := gort.NewPool(gort.WithWorkers(4))
		defer pool.Close()

		res, err := gort.RunInPool(pool, func() (int, error) {
			return winescape.Getpid()
		})
		if err != nil {
			return err
		}
		if res <= 0 {
			return fmt.Errorf("invalid pid from worker pool: %d", res)
		}
		return nil
	})

	fmt.Println("------------------------------------------------------")
	fmt.Printf("RESULTS: %d passed, %d failed\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
