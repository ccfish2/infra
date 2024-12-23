package metrics

import (
	"sync/atomic"

	"github.com/ccfish2/infra/pkg/metrics/metric"
	"github.com/sirupsen/logrus"
)

var (
	KubernetesEventProcessed = NoOpCounterVec
	DolphinAgentNamespace    = "dolphin"
	DolphinOperatorNamespace = "dolphin_operator"
)

var (
	Namespace      = DolphinAgentNamespace
	ErrorsWarnings = NoOpCounterVec
)

func InitOperatorMetrics() metric.Vec[metric.Counter] {
	return metric.NewCounterVec(metric.CounterOpts{
		ConfigName: Namespace + "_errors_warning_total",
		Namespace:  Namespace,
		Name:       "errors_warnings_total",
		Help:       "operator error and warning totalls",
	}, []string{"level", "subsystem"})
}

type LoggingHook struct {
	errs, wars atomic.Uint64
}

func NewLoggingHook() *LoggingHook {
	lh := &LoggingHook{}
	go func() {
		<-metricInitialized
		ErrorsWarnings.WithLabelValues(logrus.ErrorLevel.String(), "init").Add(float64(lh.errs.Load()))
		ErrorsWarnings.WithLabelValues(logrus.WarnLevel.String(), "init").Add(float64(lh.wars.Load()))
	}()
	return lh
}

func (lh *LoggingHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}
}
