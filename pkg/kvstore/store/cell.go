package store

type Factory interface {
	NewSyncStore(clusterName string, backend SyncStoreBackend, prefix string, opts ...WSSOpt) SyncStore
	NewWatchStore(clusterName string, keyCreator KeyCreator, observer Observer, opts ...RWSOpt) WatchStore
	NewWatchStoreManager(backend WatchStoreBackend, clusterName string) WatchStoreManager
}
