package stream

import "context"

type ToChannelOpts struct {
	bufferSize int
	errChan    chan error
}

type ToChannelOpt func(*ToChannelOpts)

func ToChannel[T any](ctx context.Context, src Observable[T], opts ...ToChannelOpt) <-chan T {
	var o ToChannelOpts
	for _, op := range opts {
		op(&o)
	}
	items := make(chan T, o.bufferSize)
	src.Observe(ctx, func(t T) { items <- t }, func(err error) {
		close(items)
		if o.errChan != nil {
			o.errChan <- err
		}
	})
	return items
}
