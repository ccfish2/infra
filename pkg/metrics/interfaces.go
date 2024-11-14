package metrics

import (
	metricpkg "github.com/ccfish2/infra/pkg/metrics/metric"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	NoOpMetric               prometheus.Metric                = &mockMetric{}
	NoOpCounterVec           metricpkg.Vec[metricpkg.Counter] = &counterVec{NoOpCollector}
	NoOpGauge                metricpkg.Gauge                  = &gauge{NoOpMetric, NoOpCollector}
	KubernetesEventProcessed                                  = NoOpCounterVec
)

type mockMetric struct{}

func (m *mockMetric) Desc() *prometheus.Desc  { return nil }
func (m *mockMetric) Write(*dto.Metric) error { return nil }

type collector struct{}

func (c *collector) Describe(chan<- *prometheus.Desc) {}
func (c *collector) Collect(chan<- prometheus.Metric) {}

type gauge struct {
	prometheus.Metric
	prometheus.Collector
}

func (g *gauge) Set(float64)          {}
func (g *gauge) Get() float64         { return 0 }
func (g *gauge) Inc()                 {}
func (g *gauge) Dec()                 {}
func (g *gauge) Add(float64)          {}
func (g *gauge) Sub(float64)          {}
func (g *gauge) SetToCurrentTime()    {}
func (g *gauge) IsEnabled() bool      { return false }
func (g *gauge) SetEnabled(bool)      {}
func (g *gauge) Opts() metricpkg.Opts { return metricpkg.Opts{} }

var (
	NoOpCollector prometheus.Collector = &collector{}
)

type counterVec struct{ prometheus.Collector }

// WithLabelValues implements metric.Vec.
func (cv *counterVec) WithLabelValues(...string) metricpkg.Counter { return NoOpGauge }
