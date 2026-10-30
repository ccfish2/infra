package job

import (
	"time"

	"github.com/ccfish2/infra/pkg/metrics"
	"github.com/ccfish2/infra/pkg/metrics/metric"
)

type jobMetrics struct {
	JobErrorsTotal      metric.Vec[metric.Counter]
	OneShotRunDuration  metric.Vec[metric.Observer]
	TimerRunDuration    metric.Vec[metric.Observer]
	ObserverRunDuration metric.Vec[metric.Observer]
}

func newJobMetrics() *jobMetrics {
	return &jobMetrics{
		JobErrorsTotal: metric.NewCounterVec(metric.CounterOpts{
			ConfigName: metrics.Namespace + "jobs_errors_total",
			Namespace:  metrics.Namespace,
			Subsystem:  "jobs",
			Name:       "errors_total",
			Help:       "The amount of errors encountered while running jobs",
		}, []string{"job"}),
		OneShotRunDuration: metric.NewHistogramVec(metric.HistogramOpts{
			ConfigName: metrics.Namespace + "jobs_one_shot_run_seconds",
			Namespace:  metrics.Namespace,
			Subsystem:  "jobs",
			Name:       "one_shot_run_seconds",
			Help:       "The run time of a one shot job",
		}, []string{"job"}),
		TimerRunDuration: metric.NewHistogramVec(metric.HistogramOpts{
			ConfigName: metrics.Namespace + "jobs_timer_run_seconds",
			Namespace:  metrics.Namespace,
			Subsystem:  "jobs",
			Name:       "timer_run_seconds",
			Help:       "The run time of a timer job",
		}, []string{"job"}),
		ObserverRunDuration: metric.NewHistogramVec(metric.HistogramOpts{
			ConfigName: metrics.Namespace + "jobs_observer_run_seconds",
			Namespace:  metrics.Namespace,
			Subsystem:  "jobs",
			Name:       "observer_run_seconds",
			Help:       "The run time of a observer job",
		}, []string{"job"}),
	}
}

type Metrics interface {
	JobError(name string, err error)
	OneShotRunDuration(name string, duration time.Duration)
	TimerRunDuration(name string, duration time.Duration)
	TimerTriggerStats(name string, latency time.Duration, folds int)
	ObserverRunDuration(name string, duration time.Duration)
}

type NopMetrics struct{}

// JobError implements Metrics.
func (NopMetrics) JobError(name string, err error) {
}

// ObserverRunDuration implements Metrics.
func (NopMetrics) ObserverRunDuration(name string, duration time.Duration) {
}

// OneShotRunDuration implements Metrics.
func (NopMetrics) OneShotRunDuration(name string, duration time.Duration) {
}

// TimerRunDuration implements Metrics.
func (NopMetrics) TimerRunDuration(name string, duration time.Duration) {
}

// TimerTriggerStats implements Metrics.
func (NopMetrics) TimerTriggerStats(name string, latency time.Duration, folds int) {
}

var _ Metrics = NopMetrics{}
