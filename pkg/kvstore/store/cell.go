package store

import "github.com/ccfish2/infra/pkg/hive/cell"

var Cell = cell.Module(
	"kvstore-utils",
	"Manage kvstore controllers",
	cell.Provide(NewFactory),
)

type Factory interface {
	NewSyncStore(clusterName string, backend SyncStoreBackend, prefix string, opts ...WSSOpt) SyncStore
	NewWatchStore(clusterName string, keyCreator KeyCreator, observer Observer, opts ...RWSOpt) WatchStore
	NewWatchStoreManager(backend WatchStoreBackend, clusterName string) WatchStoreManager
}

type factoryImpl struct {
	metrics *Metrics
}

// NewSyncStore implements Factory.
func (f *factoryImpl) NewSyncStore(clusterName string, backend SyncStoreBackend, prefix string, opts ...WSSOpt) SyncStore {
	panic("unimplemented")
}

// NewWatchStore implements Factory.
func (f *factoryImpl) NewWatchStore(clusterName string, keyCreator KeyCreator, observer Observer, opts ...RWSOpt) WatchStore {
	panic("unimplemented")
}

// NewWatchStoreManager implements Factory.
func (f *factoryImpl) NewWatchStoreManager(backend WatchStoreBackend, clusterName string) WatchStoreManager {
	panic("unimplemented")
}

func NewFactory(storeMetrics *Metrics) Factory {
	return &factoryImpl{
		metrics: storeMetrics,
	}
}
