package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func NewHistogram(opts HistogramOpts) Histogram {
	return &histogram{
		Histogram: prometheus.NewHistogram(opts.toPrometheus()),
		metric: metric{
			enabled: !opts.Disabled,
			opts:    opts.opts(),
		},
	}
}

type Histogram interface {
	prometheus.Histogram
	WithMetadata
}

type histogram struct {
	prometheus.Histogram
	metric
}

func (h *histogram) Collect(metricChan chan<- prometheus.Metric) {
	if h.enabled {
		h.Histogram.Collect(metricChan)
	}
}

func (h *histogram) Observe(val float64) {
	if h.enabled {
		h.Histogram.Observe(val)
	}
}

type Observer interface {
	prometheus.Observer
	WithMetadata
}

type observer struct {
	prometheus.Observer
	metric
}

func (o *observer) Observe(val float64) {
	if o.enabled {
		o.Observer.Observe(val)
	}
}

func NewHistogramVec(opts HistogramOpts, labelNames []string) *histogramVec {
	return &histogramVec{
		ObserverVec: prometheus.NewHistogramVec(opts.toPrometheus(), labelNames),
		metric: metric{
			enabled: !opts.Disabled,
			opts:    opts.opts(),
		},
	}
}

func NewHistogramVecWithLabels(opts HistogramOpts, labels Labels) *histogramVec {
	hv := NewHistogramVec(opts, labels.labelNames())
	initLabels(&hv.metric, labels, hv, opts.Disabled)
	return hv
}

type histogramVec struct {
	prometheus.ObserverVec
	metric
}

func (cv *histogramVec) CurryWith(labels prometheus.Labels) (Vec[Observer], error) {
	cv.checkLabels(labels)
	vec, err := cv.ObserverVec.CurryWith(labels)
	if err == nil {
		return &histogramVec{ObserverVec: vec, metric: cv.metric}, nil
	}
	return nil, err
}

func (cv *histogramVec) GetMetricWith(labels prometheus.Labels) (Observer, error) {
	if !cv.enabled {
		return &observer{
			metric: metric{enabled: false},
		}, nil
	}

	promObserver, err := cv.ObserverVec.GetMetricWith(labels)
	if err == nil {
		return &observer{
			Observer: promObserver,
			metric:   cv.metric,
		}, nil
	}
	return nil, err
}

func (cv *histogramVec) GetMetricWithLabelValues(lvs ...string) (Observer, error) {
	if !cv.enabled {
		return &observer{
			metric: metric{enabled: false},
		}, nil
	}

	promObserver, err := cv.ObserverVec.GetMetricWithLabelValues(lvs...)
	if err == nil {
		return &observer{
			Observer: promObserver,
			metric:   cv.metric,
		}, nil
	}
	return nil, err
}

func (cv *histogramVec) With(labels prometheus.Labels) Observer {
	if !cv.enabled {
		return &observer{
			metric: metric{enabled: false},
		}
	}
	cv.checkLabels(labels)

	promObserver := cv.ObserverVec.With(labels)
	return &observer{
		Observer: promObserver,
		metric:   cv.metric,
	}
}

func (cv *histogramVec) WithLabelValues(lvs ...string) Observer {
	if !cv.enabled {
		return &observer{
			metric: metric{enabled: false},
		}
	}
	cv.checkLabelValues(lvs...)

	promObserver := cv.ObserverVec.WithLabelValues(lvs...)
	return &observer{
		Observer: promObserver,
		metric:   cv.metric,
	}
}

func (cv *histogramVec) SetEnabled(e bool) {
	if !e {
		if histVec, ok := cv.ObserverVec.(*prometheus.HistogramVec); ok {
			histVec.Reset()
		}
	}

	cv.metric.SetEnabled(e)
}

type HistogramOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string

	ConstLabels prometheus.Labels

	Buckets                      []float64
	NativeHistogramBucketFactor  float64
	NativeHistogramZeroThreshold float64

	NativeHistogramMaxBucketNumber  uint32
	NativeHistogramMinResetDuration time.Duration
	NativeHistogramMaxZeroThreshold float64

	ConfigName string
	Disabled   bool
}

func (ho HistogramOpts) opts() Opts {
	return Opts{
		Namespace:   ho.Namespace,
		Subsystem:   ho.Subsystem,
		Name:        ho.Name,
		Help:        ho.Help,
		ConstLabels: ho.ConstLabels,
		ConfigName:  ho.ConfigName,
		Disabled:    ho.Disabled,
	}
}

func (ho HistogramOpts) toPrometheus() prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Namespace:                       ho.Namespace,
		Subsystem:                       ho.Subsystem,
		Name:                            ho.Name,
		Help:                            ho.Help,
		ConstLabels:                     ho.ConstLabels,
		Buckets:                         ho.Buckets,
		NativeHistogramBucketFactor:     ho.NativeHistogramBucketFactor,
		NativeHistogramZeroThreshold:    ho.NativeHistogramZeroThreshold,
		NativeHistogramMaxBucketNumber:  ho.NativeHistogramMaxBucketNumber,
		NativeHistogramMinResetDuration: ho.NativeHistogramMinResetDuration,
		NativeHistogramMaxZeroThreshold: ho.NativeHistogramMaxZeroThreshold,
	}
}
