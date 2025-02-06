package spanstat

import (
	"time"

	"github.com/ccfish2/infra/pkg/lock"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/ccfish2/infra/pkg/safetime"
)

var (
	subSystem = "spanstat"
	log       = logging.DefaultLogger.WithField(logfields.LogSubsys, subSystem)
)

// the time spent between start and stop
type SpanStat struct {
	mutex           lock.RWMutex
	spanStart       time.Time
	successDuration time.Duration
	failureDuration time.Duration
}

func Start() *SpanStat {
	s := &SpanStat{}
	return s.Start()
}

func (s *SpanStat) Start() *SpanStat {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.spanStart = time.Now()
	return s
}

func (s *SpanStat) End(success bool) *SpanStat {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.end(success)
}

func (s *SpanStat) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.successDuration = 0
	s.failureDuration = 0
}

func (s *SpanStat) end(success bool) *SpanStat {
	if !s.spanStart.IsZero() {
		d, _ := safetime.TimeSinceSafe(s.spanStart, log)
		if success {
			s.successDuration += d
		} else {
			s.failureDuration += d
		}
	}
	s.spanStart = time.Time{}
	return s
}

func (s *SpanStat) Seconds() float64 {
	if !s.spanStart.IsZero() {
		s.end(true)
	}
	tot := s.successDuration + s.failureDuration
	return tot.Seconds()
}

func (s *SpanStat) Total() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.successDuration + s.failureDuration
}
