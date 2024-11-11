package stream

import "context"

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
