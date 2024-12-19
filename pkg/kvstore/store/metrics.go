package store

import (
	"github.com/ccfish2/infra/pkg/metrics/metric"
)

type Metrics struct {
	KVStoreSyncQueueSize        metric.Vec[metric.Gauge]
	KVStoreSyncErrors           metric.Vec[metric.Counter]
	KVStoreInitialSyncCompleted metric.Vec[metric.Gauge]
}

func MetricsProvider() *Metrics {
	return &Metrics{}
}
