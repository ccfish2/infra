package main

import (
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/metrics/metric"
)

var exampleMetricCell = cell.Provide(newexampleMetric)

type exampleMetric struct {
	ExampleCounter metric.Counter
}

func newexampleMetric() exampleMetric {
	return exampleMetric{
		ExampleCounter: metric.NewCounter(metric.CounterOpts{}),
	}
}
