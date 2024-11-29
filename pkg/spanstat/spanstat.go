package spanstat

import (
	"time"

	"github.com/ccfish2/infra/pkg/lock"
)

// the time spent between start and stop
type SpanStat struct {
	mutex           lock.RWMutex
	spanStart       time.Time
	successDuration time.Duration
	failureDuration time.Duration
}

func (s *SpanStat) Start() *SpanStat {
	panic("rels")
}

func (s *SpanStat) End(success bool) *SpanStat {
	panic("rels")
}

func (s *SpanStat) Reset() {
	panic("rels")
}

func (s *SpanStat) Seconds() float64 {
	panic("release")
}
