package metric

import "github.com/prometheus/client_golang/prometheus"

type Counter interface {
	prometheus.Counter
	WithMetadata

	Get() float64
}
