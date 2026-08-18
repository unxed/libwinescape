package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	winescape "github.com/unxed/libwinescape/go"
	"github.com/unxed/libwinescape/go/gort"
)

func main() {
	fmt.Println("=== libwinescape End-to-End Integration Test Suite ===")
	fmt.Printf("Environment: IsWine=%v HostOS=%q Available=%v\n\n",
		winescape.IsWine(), winescape.HostOS(), winescape.Available())

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

	runTest("io/fs.FS DirFS and ReadFile", func() error {
		dfs := winescape.DirFS("/tmp")
		fn := filepath.Base(testFile) + ".dfs"
		full := filepath.Join("/tmp", fn)
		defer winescape.Unlink(full)

		fd, err := winescape.Open(full, winescape.O_WRONLY|winescape.O_CREAT|winescape.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		winescape.Write(fd, []byte("DirFS test content"))
		winescape.Close(fd)

		data, err := fs.ReadFile(dfs, fn)
		if err != nil {
			return fmt.Errorf("fs.ReadFile error: %w", err)
		}
		if string(data) != "DirFS test content" {
			return fmt.Errorf("data mismatch: %q", string(data))
		}
		return nil
	})

	runTest("Host POSIX Identity (getuid, getgid, getppid)", func() error {
		uid := winescape.Getuid()
		gid := winescape.Getgid()
		ppid := winescape.Getppid()
		if uid < 0 || gid < 0 || ppid < 0 {
			return fmt.Errorf("invalid identity values: uid=%d gid=%d ppid=%d", uid, gid, ppid)
		}
		fmt.Printf("[uid=%d, gid=%d, ppid=%d] ", uid, gid, ppid)
		return nil
	})

	runTest("Host Pipe2 syscall", func() error {
		r, w, err := winescape.Pipe2(winescape.O_CLOEXEC)
		if err != nil {
			return fmt.Errorf("pipe2 error: %w", err)
		}
		defer winescape.Close(r)
		defer winescape.Close(w)

		msg := []byte("ping through raw host pipe")
		nw, err := winescape.Write(w, msg)
		if err != nil || nw != len(msg) {
			return fmt.Errorf("pipe write error: %w", err)
		}

		rbuf := make([]byte, 64)
		nr, err := winescape.Read(r, rbuf)
		if err != nil {
			return fmt.Errorf("pipe read error: %w", err)
		}
		if string(rbuf[:nr]) != string(msg) {
			return fmt.Errorf("pipe data mismatch: got %q, want %q", string(rbuf[:nr]), string(msg))
		}
		return nil
	})

	runTest("Linux Inotify file event watcher", func() error {
		ifd, err := winescape.InotifyInit1(winescape.IN_CLOEXEC | winescape.IN_NONBLOCK)
		if err != nil {
			return fmt.Errorf("inotify_init1 error: %w", err)
		}
		defer winescape.Close(ifd)

		watchDir := fmt.Sprintf("/tmp/winescape_inotify_%d_%d", os.Getpid(), time.Now().UnixNano())
		if err := winescape.Mkdir(watchDir, 0755); err != nil {
			return fmt.Errorf("mkdir watchDir error: %w", err)
		}
		defer winescape.Rmdir(watchDir)

		wd, err := winescape.InotifyAddWatch(ifd, watchDir, winescape.IN_CREATE)
		if err != nil {
			return fmt.Errorf("inotify_add_watch error: %w", err)
		}
		defer winescape.InotifyRmWatch(ifd, wd)

		testCreated := filepath.Join(watchDir, "event.tmp")
		cfd, err := winescape.Open(testCreated, winescape.O_WRONLY|winescape.O_CREAT|winescape.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("create watched file error: %w", err)
		}
		winescape.Close(cfd)
		defer winescape.Unlink(testCreated)

		buf := make([]byte, 4096)
		n, err := winescape.Read(ifd, buf)
		if err != nil {
			return fmt.Errorf("read inotify events error: %w", err)
		}
		events, err := winescape.ParseInotifyEvents(buf[:n])
		if err != nil {
			return fmt.Errorf("ParseInotifyEvents error: %w", err)
		}

		found := false
		for _, e := range events {
			if e.Name == "event.tmp" && (e.Mask&winescape.IN_CREATE != 0) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected IN_CREATE event for 'event.tmp', got: %+v", events)
		}
		return nil
	})

	runTest("AF_UNIX socket IPC (DialUnix & ListenUnix)", func() error {
		sockPath := fmt.Sprintf("/tmp/winescape_ipc_%d_%d.sock", os.Getpid(), time.Now().UnixNano())
		lfd, err := winescape.ListenUnix(sockPath, 5)
		if err != nil {
			return fmt.Errorf("ListenUnix error: %w", err)
		}
		defer winescape.Close(lfd)
		defer winescape.Unlink(sockPath)

		errChan := make(chan error, 1)
		msg := []byte("hello directly over raw host AF_UNIX socket!")

		go func() {
			nfd, err := winescape.Accept4(lfd, nil, nil, winescape.SOCK_CLOEXEC)
			if err != nil {
				errChan <- fmt.Errorf("Accept4 error: %w", err)
				return
			}
			defer winescape.Close(nfd)

			buf := make([]byte, 128)
			n, err := winescape.Read(nfd, buf)
			if err != nil {
				errChan <- fmt.Errorf("server read error: %w", err)
				return
			}
			if string(buf[:n]) != string(msg) {
				errChan <- fmt.Errorf("server received mismatch: got %q, want %q", string(buf[:n]), string(msg))
				return
			}
			errChan <- nil
		}()

		client, err := winescape.DialUnix(sockPath)
		if err != nil {
			return fmt.Errorf("DialUnix error: %w", err)
		}
		defer client.Close()

		if _, err := client.Write(msg); err != nil {
			return fmt.Errorf("client write error: %w", err)
		}

		select {
		case err := <-errChan:
			if err != nil {
				return err
			}
		case <-time.After(2 * time.Second):
			return fmt.Errorf("timeout waiting for AF_UNIX IPC server")
		}

		return nil
	})

	runTest("Host Mmap / Munmap", func() error {
		fn := testFile + ".mmap"
		defer winescape.Unlink(fn)

		fd, err := winescape.Open(fn, winescape.O_RDWR|winescape.O_CREAT|winescape.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer winescape.Close(fd)

		msg := []byte("mmap test raw payload")
		winescape.Write(fd, msg)

		b, err := winescape.Mmap(fd, 0, len(msg), winescape.PROT_READ, winescape.MAP_SHARED)
		if err != nil {
			return fmt.Errorf("mmap error: %w", err)
		}
		if string(b) != string(msg) {
			winescape.Munmap(b)
			return fmt.Errorf("mmap content mismatch: %q", string(b))
		}
		if err := winescape.Munmap(b); err != nil {
			return fmt.Errorf("munmap error: %w", err)
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
