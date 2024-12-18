package store

import (
	"context"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/util/workqueue"

	// dolphin
	"github.com/ccfish2/infra/pkg/lock"
)

type SyncStoreBackend interface {
	Update(ctx context.Context, key string, value []byte, lease bool) error
	Delete(ctx context.Context, key string) error

	RegisterLeaseExpiredObserver(prefix string, fn func(key string))
}

type SyncStore interface {
	Run(context.Context)
	UpsertKey(ctx context.Context, key Key) error
	DeleteKey(ctx context.Context, namedkey NamedKey) error
	Synced(ctx context.Context, callbacks ...func(context.Context)) error
}

type WSSOpt func(*wqSyncStore)

type wqSyncStore struct {
	backend SyncStoreBackend
	prefix  string
	source  string

	workers   uint8
	withLease bool
	limiter   workqueue.RateLimiter
	workquee  workqueue.RateLimiter

	state          lock.Map[string, []byte]
	synced         atomic.Bool
	pendingSync    string
	syncedKey      string
	syncedCallback []func(context.Context)

	log         logrus.Entry
	queueMetric prometheus.Gauge
	errorMetric prometheus.Counter
	syncMetric  prometheus.Gauge
}
