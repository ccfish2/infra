package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	errClosed  = errors.New("workerpool closed")
	errDrained = errors.New("workerpool drained")
)

type WorkerPool struct {
	workers chan struct{}
	tasks   chan *task
	cancel  context.CancelFunc
	results []Task
	wg      sync.WaitGroup

	mu       sync.Mutex
	draining bool
	closed   bool
}

func New(workers int) *WorkerPool {
	return NewWithContext(context.Background(), workers)
}

func NewWithContext(ctx context.Context, workers int) *WorkerPool {
	if workers < 0 {
		panic(fmt.Errorf("workerpool: workers must be >= 0, got %d", workers))
	}
	wp := &WorkerPool{
		workers: make(chan struct{}, workers),
		tasks:   make(chan *task),
	}
	ctx, cancel := context.WithCancel(ctx)
	wp.cancel = cancel
	go wp.run(ctx)
	return wp
}

func (wp *WorkerPool) Cap() int {
	return cap(wp.workers)
}

func (wp *WorkerPool) Len() int {
	return len(wp.workers)
}

func (wp *WorkerPool) Submit(id string, f func(context.Context) error) error {
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		return errClosed
	}
	if wp.draining {
		wp.mu.Unlock()
		return errDrained
	}
	wp.wg.Add(1)
	wp.mu.Unlock()
	wp.tasks <- &task{id: id, run: f}
	return nil
}

func (wp *WorkerPool) Drain() ([]Task, error) {
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		return nil, errClosed
	}
	if wp.draining {
		wp.mu.Unlock()
		return nil, errDrained
	}
	wp.draining = true
	wp.mu.Unlock()

	wp.wg.Wait()

	res := wp.results
	wp.results = nil

	wp.mu.Lock()
	wp.draining = false
	wp.mu.Unlock()

	return res, nil
}

func (wp *WorkerPool) Close() error {
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		return errClosed
	}
	wp.closed = true
	wp.mu.Unlock()

	wp.cancel()
	close(wp.tasks)
	wp.wg.Wait()

	<-wp.workers
	return nil
}
func (wp *WorkerPool) run(ctx context.Context) {
	for t := range wp.tasks {
		t := t
		if t.run == nil {
			continue
		}
		wp.workers <- struct{}{}
		go func() {
			defer wp.wg.Done()
			result := taskResult{id: t.id}
			result.err = t.run(ctx)
			wp.mu.Lock()
			wp.results = append(wp.results, &result)
			wp.mu.Unlock()

			<-wp.workers
		}()

	}
	close(wp.workers)
}
