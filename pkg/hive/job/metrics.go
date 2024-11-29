package job

import "github.com/ccfish2/infra/pkg/metrics/metric"

type jobMetrics struct {
	JobErrorsTotal      metric.Vec[metric.Counter]
	OneShotRunDuration  metric.Vec[metric.Observer]
	TimerRunDuration    metric.Vec[metric.Observer]
	ObserverRunDuration metric.Vec[metric.Observer]
}
