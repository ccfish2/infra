package stream

import "context"

type Observable[T any] interface {
	Observe(ctx context.Context, next func(T), complete func(error))
}

type FuncObservable[T any] func(context.Context, func(T), func(error))

func (f FuncObservable[T]) Observe(ctx context.Context, next func(T), complete func(error)) {
	f(ctx, next, complete)
}
