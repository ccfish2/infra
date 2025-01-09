package rate

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
)

/*similar as rate limiter*/
type Limiter struct {
	semaphore   *semaphore.Weighted
	burst       int64
	currWeights atomic.Int64
	ticker      *time.Ticker
	cancelFunc  context.CancelFunc
	ctx         context.Context
}

func NewLimiter(interval time.Duration, b int64) *Limiter {
	ticker := time.NewTicker(interval)
	ctx, cancel := context.WithCancel(context.Background())
	l := &Limiter{
		ticker:     ticker,
		burst:      b,
		semaphore:  semaphore.NewWeighted(b),
		cancelFunc: cancel,
		ctx:        ctx,
	}

	go func() {
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			cur := l.currWeights.Swap(0)
			l.semaphore.Release(cur)
		}
	}()
	return l
}

func (lim *Limiter) Stop() {
	lim.cancelFunc()
	lim.ticker.Stop()
}

func (lim *Limiter) assertLive() {
	select {
	case <-lim.ctx.Done():
		panic("should not stop")
	default:
	}
}

func (lim *Limiter) Allow() bool {
	return lim.AllowN(1)
}

func (lim *Limiter) AllowN(n int64) bool {
	lim.assertLive()
	acq := lim.semaphore.TryAcquire(n)
	if acq {
		lim.currWeights.Add(n)
		return true
	}
	return false
}

func (lim *Limiter) Wait(ctx context.Context) error {
	return lim.WaitN(ctx, 1)
}

func (lim *Limiter) WaitN(ctx context.Context, n int64) error {
	lim.assertLive()
	if n > lim.burst {
		return fmt.Errorf("out of burst")
	}
	err := lim.semaphore.Acquire(ctx, n)
	if err != nil {
		return fmt.Errorf("FAILED err %v", err)
	}
	lim.currWeights.Add(n)
	return nil
}
