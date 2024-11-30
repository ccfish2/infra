package stream

import (
	"context"

	"github.com/ccfish2/infra/pkg/lock"
)

func FromChannel[T any](in <-chan T) Observable[T] {
	return FuncObservable[T](
		func(ctx context.Context, next func(T), complete func(error)) {
			go func() {
				done := ctx.Done()
				for {
					select {
					case <-done:
						complete(ctx.Err())
						return
					case v, ok := <-in:
						if !ok {
							complete(nil)
							return
						}
						next(v)
					}
				}
			}()
		})
}

type mcastSubscriber[T any] struct {
	next     func(T)
	complete func()
}

type MulticastOpt func(o *mcastOpts)

type mcastOpts struct {
	emitLatest bool
}

func (o mcastOpts) apply(opts []MulticastOpt) mcastOpts {
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

var (
	// Emit the latest seen item when subscribing.
	EmitLatest = func(o *mcastOpts) { o.emitLatest = true }
)

func Multicast[T any](opts ...MulticastOpt) (mcast Observable[T], next func(T), complete func(error)) {
	var (
		mu          lock.Mutex
		subId       int
		subs        = make(map[int]mcastSubscriber[T])
		latestValue T
		completed   bool
		completeErr error
		haveLatest  bool
		opt         = mcastOpts{}.apply(opts)
	)

	next = func(item T) {
		mu.Lock()
		defer mu.Unlock()
		if completed {
			return
		}
		if opt.emitLatest {
			latestValue = item
			haveLatest = true
		}
		for _, sub := range subs {
			sub.next(item)
		}
	}

	complete = func(err error) {
		mu.Lock()
		defer mu.Unlock()
		completed = true
		completeErr = err
		for _, sub := range subs {
			sub.complete()
		}
		subs = nil
	}

	mcast = FuncObservable[T](
		func(ctx context.Context, subNext func(T), subComplete func(error)) {
			mu.Lock()
			if completed {
				mu.Unlock()
				go subComplete(completeErr)
				return
			}

			subCtx, cancel := context.WithCancel(ctx)
			thisId := subId
			subId++
			subs[thisId] = mcastSubscriber[T]{
				subNext,
				cancel,
			}

			// Continue subscribing asynchronously so caller is not blocked.
			go func() {
				if opt.emitLatest && haveLatest {
					subNext(latestValue)
				}
				mu.Unlock()

				// Wait for cancellation by observer, or completion from upstream.
				<-subCtx.Done()

				// Remove the observer and complete.
				var err error
				mu.Lock()
				delete(subs, thisId)
				if completed {
					err = completeErr
				} else {
					err = subCtx.Err()
				}
				mu.Unlock()
				subComplete(err)
			}()
		})

	return
}
