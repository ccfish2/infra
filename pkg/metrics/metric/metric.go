package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type WithMetadata interface {
	IsEnabled() bool
	SetEnabled(bool)
	Opts() Opts
}

type Opts struct {
	Namespace string
	Subsystem string
	Name      string

	Help        string
	ConstLabels prometheus.Labels

	ConfigName string

	Disabled bool
}

type Vec[T any] interface {
	WithLabelValues(lvs ...string) T
}

type Gauge interface {
	prometheus.Gauge
	WithMetadata

	Get() float64
}

type Label struct {
	Name string
	// If defined, only these values are allowed.
	Values Values
}

// Values is a distinct set of possible label values for a particular Label.
type Values map[string]struct{}

// NewValues constructs a Values type from a set of strings.
func NewValues(vs ...string) Values {
	vals := Values{}
	for _, v := range vs {
		vals[v] = struct{}{}
	}
	return vals
}

// Labels is a slice of labels that represents a label set for a vector type
// metric.
type Labels []Label
type labelSet struct {
	lbls Labels
	m    map[string]map[string]struct{}
}

type metric struct {
	enabled bool
	opts    Opts
	labels  *labelSet
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

// Set sets the Gauge to an arbitrary value.
func (g *gauge) Set(val float64) {
	if g.enabled {
		g.Gauge.Set(val)
	}
}

// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
// values.
func (g *gauge) Inc() {
	if g.enabled {
		g.Gauge.Inc()
	}
}

// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
// values.
func (g *gauge) Dec() {
	if g.enabled {
		g.Gauge.Dec()
	}
}

// Add adds the given value to the Gauge. (The value can be negative,
// resulting in a decrease of the Gauge.)
func (g *gauge) Add(val float64) {
	if g.enabled {
		g.Gauge.Add(val)
	}
}

// Sub subtracts the given value from the Gauge. (The value can be
// negative, resulting in an increase of the Gauge.)
func (g *gauge) Sub(i float64) {
	if g.enabled {
		g.Gauge.Sub(i)
	}
}

// SetToCurrentTime sets the Gauge to the current Unix time in seconds.
func (g *gauge) SetToCurrentTime() {
	if g.enabled {
		g.Gauge.SetToCurrentTime()
	}
}
