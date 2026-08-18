package gort

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_BasicExecution(t *testing.T) {
	pool := NewPool(WithWorkers(2), WithQueueSize(4))
	defer pool.Close()

	var counter int64
	err := pool.Do(func() {
		atomic.AddInt64(&counter, 1)
	})
	if err != nil {
		t.Fatalf("unexpected Do error: %v", err)
	}

	if atomic.LoadInt64(&counter) != 1 {
		t.Errorf("expected counter 1, got %d", counter)
	}
}

func TestPool_ConcurrentExecution(t *testing.T) {
	pool := NewPool(WithWorkers(4))
	defer pool.Close()

	const tasks = 50
	var counter int64
	var wg sync.WaitGroup
	wg.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func() {
			defer wg.Done()
			err := pool.Do(func() {
				atomic.AddInt64(&counter, 1)
			})
			if err != nil {
				t.Errorf("pool.Do failed: %v", err)
			}
		}()
	}

	wg.Wait()
	if atomic.LoadInt64(&counter) != tasks {
		t.Errorf("expected counter %d, got %d", tasks, counter)
	}
}

func TestPool_ContextCancel(t *testing.T) {
	pool := NewPool(WithWorkers(1), WithQueueSize(1))
	defer pool.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := pool.DoContext(ctx, func() {
		t.Errorf("cancelled task must not execute")
	})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestPool_Close(t *testing.T) {
	pool := NewPool(WithWorkers(2))
	pool.Close()

	err := pool.Do(func() {})
	if err != ErrPoolClosed {
		t.Errorf("expected ErrPoolClosed, got %v", err)
	}

	// Idempotent Close
	pool.Close()
}

func TestRun_OneOff(t *testing.T) {
	var executed bool
	Run(func() {
		executed = true
	})
	if !executed {
		t.Errorf("Run did not execute function")
	}
}

func TestRunInPool_Generic(t *testing.T) {
	pool := NewPool(WithWorkers(2))
	defer pool.Close()

	res, err := RunInPool(pool, func() (int, error) {
		time.Sleep(5 * time.Millisecond)
		return 42, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 42 {
		t.Errorf("expected 42, got %d", res)
	}
}
