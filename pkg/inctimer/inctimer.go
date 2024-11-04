package inctimer

import "time"

type IncTimer interface {
	After(time.Duration) <-chan time.Time
}

type incTimer struct {
	t *time.Timer
}

func New() (IncTimer, func() bool) {
	it := &incTimer{}
	return it, it.stop
}

func (it *incTimer) stop() bool {
	if it.t == nil {
		return false
	}
	return it.t.Stop()
}

func (it *incTimer) After(d time.Duration) <-chan time.Time {
	it.stop()
	it.t = time.NewTimer(d)
	return it.t.C
}
