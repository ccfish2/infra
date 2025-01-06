package promise

import (
	"context"
	"sync"
)

const (
	promiseUnresolved = iota
	promiseResolved
	promiseRejected
)

type Promise[T any] interface {
	Await(context.Context) (T, error)
}

type Resolver[T any] interface {
	Resolve(T)
	Reject(error)
}

type promise[T any] struct {
	mu    sync.Mutex
	cond  *sync.Cond
	state int
	value T
	err   error
}

// Lock implements sync.Locker.
func (p *promise[T]) Lock() {
	p.mu.Lock()
}

// Unlock implements sync.Locker.
func (p *promise[T]) Unlock() {
	p.mu.Unlock()
}

func New[T any]() (Resolver[T], Promise[T]) {
	promise := &promise[T]{}
	promise.cond = sync.NewCond(promise)
	return promise, promise
}

func (p *promise[T]) Resolve(value T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != promiseUnresolved {
		return
	}
	p.state = promiseResolved
	p.value = value
	p.cond.Broadcast()
}

func (p *promise[T]) Reject(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == promiseRejected {
		return
	}
	p.state = promiseRejected
	p.err = err
	p.cond.Broadcast()
}

func (p *promise[T]) Await(ctx context.Context) (value T, err error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go func() {
		<-ctx.Done()
		p.cond.Broadcast()
	}()

	p.mu.Lock()
	defer p.mu.Unlock()

	for (p.state == promiseUnresolved) && (ctx == nil || ctx.Err() == nil) {
		p.cond.Wait()
	}

	if ctx.Err() != nil {
		err = ctx.Err()
	} else if p.state == promiseResolved {
		value = p.value
	} else {
		err = p.err
	}
	return
}

type wrappedPromise[T any] func(context.Context) (T, error)

// Await implements Promise.
func (w wrappedPromise[T]) Await(ctx context.Context) (T, error) {
	return w(ctx)
}

func MapError[T any](p Promise[T], transform func(error) error) Promise[T] {
	return wrappedPromise[T](func(ctx context.Context) (out T, err error) {
		v, err := p.Await(ctx)
		if err != nil {
			err = transform(err)
		}
		return v, err
	})
}
