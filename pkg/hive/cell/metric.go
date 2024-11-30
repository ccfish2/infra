package cell

import (
	"fmt"
	"reflect"

	"github.com/ccfish2/infra/pkg/hive/internal"
	pkgmetric "github.com/ccfish2/infra/pkg/metrics/metric"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/dig"
)

var (
	withMeta  pkgmetric.WithMetadata
	collector prometheus.Collector
)

func Metric[S any](ctor func() S) Cell {
	var outNil S
	outType := reflect.TypeOf(outNil)
	if outType.Kind() == reflect.Ptr {
		outType = outType.Elem()
	}

	if outType.Kind() != reflect.Struct {
		panic("it must be a struct type")
	}

	if outType.NumField() == 0 {
		panic("the struct must have fields")
	}
	metaTyp := reflect.TypeOf(&withMeta).Elem()
	collectTyp := reflect.TypeOf(&collector).Elem()
	for i := 0; i < outType.NumField(); i++ {
		field := outType.Field(i)
		if !field.IsExported() {
			panic("the field does not exposed")
		}
		if !field.Type.Implements(metaTyp) {
			panic("field does not implement meta type")
		}

		if !field.Type.Implements(collectTyp) {
			panic("field does not implement collector type")
		}
	}
	return &metric[S]{
		ctor: ctor,
	}
}

type metric[S any] struct {
	ctor func() S
}

type metricOut struct {
	dig.Out

	Metrics []pkgmetric.WithMetadata `group:"hive-metrics,flatten"`
}

func (m *metric[S]) provideMetrics(metricSet S) metricOut {
	var metrics []pkgmetric.WithMetadata
	value := reflect.ValueOf(metrics)
	typ := value.Type()
	if typ.Kind() == reflect.Pointer {
		value = value.Elem()
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return metricOut{}
	}

	for i := 0; i < typ.NumField(); i++ {
		withmeta, ok := value.Field(i).Interface().(pkgmetric.WithMetadata)
		if ok {
			metrics = append(metrics, withmeta)
		}
	}
	return metricOut{
		Metrics: metrics,
	}
}

// Apply implements Cell.
func (m *metric[S]) Apply(container container) error {
	container.Provide(m.ctor, dig.Export(true))
	container.Provide(m.provideMetrics, dig.Export(true))
	return nil
}

// Info implements Cell.
func (m *metric[S]) Info(container) Info {
	n := NewInfoNode(fmt.Sprintf("📈 %s", internal.FuncNameAndLocation(m.ctor)))
	n.condensed = true

	return n
}
