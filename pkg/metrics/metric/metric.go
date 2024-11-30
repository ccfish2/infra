package metric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"

	"github.com/ccfish2/infra/pkg/logging/logfields"
	collections "github.com/ccfish2/infra/pkg/metrics/metric/collection"
)

var logger = logrus.WithField(logfields.LogSubsys, "metric")

type WithMetadata interface {
	IsEnabled() bool
	SetEnabled(bool)
	Opts() Opts
}

type metric struct {
	enabled bool
	opts    Opts
	labels  *labelSet
}

func (b *metric) forEachLabelVector(fn func(lvls []string)) {
	if b.labels == nil {
		return
	}
	var labelValues [][]string
	for _, label := range b.labels.lbls {
		labelValues = append(labelValues, maps.Keys(label.Values))
	}
	for _, labelVector := range collections.CartesianProduct(labelValues...) {
		fn(labelVector)
	}
}

func (b *metric) checkLabelValues(lvs ...string) {
	if b.labels == nil {
		return
	}
	if err := b.labels.checkLabelValues(lvs); err != nil {
		logger.WithError(err).
			WithFields(logrus.Fields{
				"metric": b.opts.Name,
			}).
			Warning("metric label constraints violated, metric will still be collected")
	}
}

func (b *metric) checkLabels(labels prometheus.Labels) {
	if b.labels == nil {
		return
	}

	if err := b.labels.checkLabels(labels); err != nil {
		logger.WithError(err).
			WithFields(logrus.Fields{
				"metric": b.opts.Name,
			}).
			Warning("metric label constraints violated, metric will still be collected")
	}
}

func (b *metric) IsEnabled() bool {
	return b.enabled
}

func (b *metric) SetEnabled(e bool) {
	b.enabled = e
}

func (b *metric) Opts() Opts {
	return b.opts
}

type Vec[T any] interface {
	prometheus.Collector
	WithMetadata
	CurryWith(labels prometheus.Labels) (Vec[T], error)
	GetMetricWith(labels prometheus.Labels) (T, error)

	GetMetricWithLabelValues(lvs ...string) (T, error)

	With(labels prometheus.Labels) T

	WithLabelValues(lvs ...string) T
}

type DeletableVec[T any] interface {
	Vec[T]

	Delete(labels prometheus.Labels) bool
	DeleteLabelValues(lvs ...string) bool
	DeletePartialMatch(labels prometheus.Labels) int
	Reset()
}
type Opts struct {
	Namespace   string
	Subsystem   string
	Name        string
	Help        string
	ConstLabels prometheus.Labels
	ConfigName  string
	Disabled    bool
}

func (b Opts) GetConfigName() string {
	if b.ConfigName == "" {
		return prometheus.BuildFQName(b.Namespace, b.Subsystem, b.Name)
	}
	return b.ConfigName
}

type Label struct {
	Name   string
	Values Values
}
type Values map[string]struct{}

func NewValues(vs ...string) Values {
	vals := Values{}
	for _, v := range vs {
		vals[v] = struct{}{}
	}
	return vals
}

type Labels []Label

func (lbls Labels) labelNames() []string {
	lns := make([]string, len(lbls))
	for i, label := range lbls {
		lns[i] = label.Name
	}
	return lns
}

type labelSet struct {
	lbls Labels
	m    map[string]map[string]struct{}
}

func (l *labelSet) namesToValues() map[string]map[string]struct{} {
	if l.m != nil {
		return l.m
	}
	l.m = make(map[string]map[string]struct{})
	for _, label := range l.lbls {
		l.m[label.Name] = label.Values
	}
	return l.m
}

func (l *labelSet) checkLabels(labels prometheus.Labels) error {
	for name, value := range labels {
		if lvs, ok := l.namesToValues()[name]; ok {
			if _, ok := lvs[value]; !ok {
				return fmt.Errorf("unexpected label vector value for label %q: value %q not defined in label range %v",
					name, value, maps.Keys(lvs))
			}
		} else {
			return fmt.Errorf("invalid label name: %s", name)
		}
	}
	return nil
}

func (l *labelSet) checkLabelValues(lvs []string) error {
	if len(l.lbls) != len(lvs) {
		return fmt.Errorf("unexpected label vector length: expected %d, got %d", len(l.lbls), len(lvs))
	}
	for i, label := range l.lbls {
		if _, ok := label.Values[lvs[i]]; !ok {
			return fmt.Errorf("unexpected label vector value for label %q: value %q not defined in label range %v",
				label.Name, lvs[i], maps.Keys(label.Values))
		}
	}
	return nil
}

func initLabels[T any](m *metric, labels Labels, vec Vec[T], disabled bool) {
	if disabled {
		return
	}
	m.labels = &labelSet{lbls: labels}
	m.forEachLabelVector(func(vs []string) {
		vec.WithLabelValues(vs...)
	})
}
