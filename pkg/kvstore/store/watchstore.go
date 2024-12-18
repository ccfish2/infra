package store

import (
	"context"
	"sync/atomic"

	"github.com/ccfish2/infra/pkg/metrics/metric"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type WatchStore interface {
	Watch(ctx context.Context, backstore WatchStoreBackend, prefix string)
	NumEntries() uint64
	Synced() bool
	Drain()
}

type WatchStoreBackend interface {
	ListAndWatch(context.Context, string, int)
}

type RWSOpt func(*restartableWatchStore)
type rwsEntry struct {
	Key   string
	State bool
}

type restartableWatchStore struct {
	source     string
	keyCreator KeyCreator
	observer   Observer

	watching        atomic.Bool
	synced          atomic.Bool
	onSyncCallbacks []func(ctx context.Context)

	state      map[string]*rwsEntry
	numEntries atomic.Uint64

	log           *logrus.Entry
	entriesMetric prometheus.Gauge
	syncMetric    metric.Vec[metric.Gauge]
}
