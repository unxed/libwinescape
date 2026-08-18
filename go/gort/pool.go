package gort

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	// ErrPoolClosed is returned when submitting tasks to a closed pool.
	ErrPoolClosed = errors.New("gort: worker pool is closed")
)

// Task represents a function executed on a dedicated, OS-thread-locked worker.
type Task func() error

type poolTask struct {
	fn  func()
	ctx context.Context
}

// PoolOption configures a worker pool.
type PoolOption func(*poolConfig)

type poolConfig struct {
	workers   int
	queueSize int
}

// WithWorkers sets the number of dedicated OS-thread-locked worker goroutines.
func WithWorkers(n int) PoolOption {
	return func(c *poolConfig) {
		if n > 0 {
			c.workers = n
		}
	}
}

// WithQueueSize sets the capacity of the task queue channel.
func WithQueueSize(size int) PoolOption {
	return func(c *poolConfig) {
		if size >= 0 {
			c.queueSize = size
		}
	}
}

// Pool manages a fixed pool of worker goroutines, each bound to an OS thread via runtime.LockOSThread.
type Pool struct {
	tasks   chan poolTask
	workers int
	closed  atomic.Bool
	wg      sync.WaitGroup
}

// NewPool creates a new worker pool with OS-thread-locked workers.
func NewPool(opts ...PoolOption) *Pool {
	cfg := poolConfig{
		workers:   runtime.NumCPU(),
		queueSize: 64,
	}
	if cfg.workers < 4 {
		cfg.workers = 4
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	p := &Pool{
		tasks:   make(chan poolTask, cfg.queueSize),
		workers: cfg.workers,
	}

	p.wg.Add(cfg.workers)
	for i := 0; i < cfg.workers; i++ {
		go p.workerLoop()
	}

	return p
}

func (p *Pool) workerLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer p.wg.Done()

	for task := range p.tasks {
		if task.ctx != nil && task.ctx.Err() != nil {
			task.fn()
			continue
		}
		task.fn()
	}
}

// Do executes fn on an OS-thread-locked worker in the pool and waits for it to complete.
func (p *Pool) Do(fn func()) error {
	return p.DoContext(context.Background(), fn)
}

// DoContext executes fn with context cancellation support.
func (p *Pool) DoContext(ctx context.Context, fn func()) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}

	done := make(chan struct{})
	t := poolTask{
		ctx: ctx,
		fn: func() {
			defer close(done)
			if ctx.Err() == nil {
				fn()
			}
		},
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.tasks <- t:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// Close gracefully closes the pool, waiting for all currently queued tasks to complete.
func (p *Pool) Close() {
	if p.closed.CompareAndSwap(false, true) {
		close(p.tasks)
		p.wg.Wait()
	}
}

// Workers returns the number of workers in the pool.
func (p *Pool) Workers() int {
	return p.workers
}

// Run runs fn on a dedicated, temporary locked OS thread without needing a persistent pool.
func Run(fn func()) {
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(done)
		fn()
	}()
	<-done
}

// RunInPool executes a typed function returning (T, error) on the provided pool.
func RunInPool[T any](p *Pool, fn func() (T, error)) (T, error) {
	var result T
	var err error
	execErr := p.Do(func() {
		result, err = fn()
	})
	if execErr != nil {
		var zero T
		return zero, execErr
	}
	return result, err
}
