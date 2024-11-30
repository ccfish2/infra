package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func NewGauge(opts GaugeOpts) Gauge {
	return &gauge{
		Gauge: prometheus.NewGauge(opts.toPrometheus()),
		metric: metric{
			enabled: !opts.Disabled,
			opts:    Opts(opts),
		},
	}
}

type Gauge interface {
	prometheus.Gauge
	WithMetadata

	Get() float64
}

type gauge struct {
	prometheus.Gauge
	metric
}

func (g *gauge) Collect(metricChan chan<- prometheus.Metric) {
	if g.enabled {
		g.Gauge.Collect(metricChan)
	}
}

func (g *gauge) Get() float64 {
	if !g.enabled {
		return 0
	}

	var pm dto.Metric
	err := g.Gauge.Write(&pm)
	if err == nil {
		return *pm.Gauge.Value
	}
	return 0
}

func (g *gauge) Set(val float64) {
	if g.enabled {
		g.Gauge.Set(val)
	}
}

func (g *gauge) Inc() {
	if g.enabled {
		g.Gauge.Inc()
	}
}

func (g *gauge) Dec() {
	if g.enabled {
		g.Gauge.Dec()
	}
}

func (g *gauge) Add(val float64) {
	if g.enabled {
		g.Gauge.Add(val)
	}
}

func (g *gauge) Sub(i float64) {
	if g.enabled {
		g.Gauge.Sub(i)
	}
}

func (g *gauge) SetToCurrentTime() {
	if g.enabled {
		g.Gauge.SetToCurrentTime()
	}
}

func NewGaugeVec(opts GaugeOpts, labelNames []string) *gaugeVec {
	gv := &gaugeVec{
		GaugeVec: prometheus.NewGaugeVec(opts.toPrometheus(), labelNames),
		metric: metric{
			enabled: !opts.Disabled,
			opts:    Opts(opts),
		},
	}
	return gv
}

func NewGaugeVecWithLabels(opts GaugeOpts, labels Labels) *gaugeVec {
	gv := NewGaugeVec(opts, labels.labelNames())
	initLabels[Gauge](&gv.metric, labels, gv, opts.Disabled)
	return gv
}

type gaugeVec struct {
	*prometheus.GaugeVec
	metric
}

func (gv *gaugeVec) CurryWith(labels prometheus.Labels) (Vec[Gauge], error) {
	gv.checkLabels(labels)
	vec, err := gv.GaugeVec.CurryWith(labels)
	if err == nil {
		return &gaugeVec{GaugeVec: vec, metric: gv.metric}, nil
	}
	return nil, err
}

func (gv *gaugeVec) GetMetricWith(labels prometheus.Labels) (Gauge, error) {
	if !gv.enabled {
		return &gauge{
			metric: metric{enabled: false},
		}, nil
	}

	promGauge, err := gv.GaugeVec.GetMetricWith(labels)
	if err == nil {
		return &gauge{
			Gauge:  promGauge,
			metric: gv.metric,
		}, nil
	}
	return nil, err
}

func (gv *gaugeVec) GetMetricWithLabelValues(lvs ...string) (Gauge, error) {
	if !gv.enabled {
		return &gauge{
			metric: metric{enabled: false},
		}, nil
	}

	promGauge, err := gv.GaugeVec.GetMetricWithLabelValues(lvs...)
	if err == nil {
		return &gauge{
			Gauge:  promGauge,
			metric: gv.metric,
		}, nil
	}
	return nil, err
}

func (gv *gaugeVec) With(labels prometheus.Labels) Gauge {
	if !gv.enabled {
		return &gauge{
			metric: metric{enabled: false},
		}
	}
	gv.checkLabels(labels)

	promGauge := gv.GaugeVec.With(labels)
	return &gauge{
		Gauge:  promGauge,
		metric: gv.metric,
	}
}

func (gv *gaugeVec) WithLabelValues(lvs ...string) Gauge {
	gv.checkLabelValues(lvs...)
	if !gv.enabled {
		return &gauge{
			metric: metric{enabled: false},
		}
	}

	promGauge := gv.GaugeVec.WithLabelValues(lvs...)
	return &gauge{
		Gauge:  promGauge,
		metric: gv.metric,
	}
}

func (gv *gaugeVec) SetEnabled(e bool) {
	if !e {
		gv.Reset()
	}

	gv.metric.SetEnabled(e)
}

type GaugeFunc interface {
	prometheus.GaugeFunc
	WithMetadata
}

func NewGaugeFunc(opts GaugeOpts, function func() float64) GaugeFunc {
	return &gaugeFunc{
		GaugeFunc: prometheus.NewGaugeFunc(opts.toPrometheus(), function),
		metric: metric{
			enabled: !opts.Disabled,
			opts:    Opts(opts),
		},
	}
}

type gaugeFunc struct {
	prometheus.GaugeFunc
	metric
}

func (gf *gaugeFunc) Collect(metricChan chan<- prometheus.Metric) {
	if gf.enabled {
		gf.GaugeFunc.Collect(metricChan)
	}
}

type GaugeOpts Opts

func (o GaugeOpts) toPrometheus() prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
	}
}
