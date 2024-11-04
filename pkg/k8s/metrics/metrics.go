package metrics

import (
	"sync"
	"time"
)

var (
	LastInteraction        eventTimestamper
	LastSuccessInteraction eventTimestamper
)

type eventTimestamper struct {
	timestamp time.Time
	lock      sync.Mutex
}

func (e *eventTimestamper) Reset() {
	e.lock.Lock()
	e.timestamp = time.Now()
	e.lock.Unlock()
}

func (e *eventTimestamper) Time() time.Time {
	e.lock.Lock()
	t := e.timestamp
	e.lock.Unlock()
	return t
}
