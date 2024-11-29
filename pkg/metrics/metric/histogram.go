package metric

import "github.com/prometheus/client_golang/prometheus"

type Observer interface {
	prometheus.Observer
	WithMetadata
}

type Histogram interface {
	prometheus.Histogram
	WithMetadata
}

// Histogram

type histogram struct {
	prometheus.Collector
}
