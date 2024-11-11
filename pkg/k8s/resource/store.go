package resource

import (
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type typedStore[T k8sRuntime.Object] struct {
	store   cache.Indexer
	release func()
}

/* similar as cache.Store */
type Store[T k8sRuntime.Object] interface {
	List() []T
	IterKeys() KeyIter

	Get(obj T) (item T, exists bool, err error)
	GetByKey(key Key) (item T, exists bool, err error)
	IndexKeys(indexName, indexedValue string) ([]string, error)
	ByIndex(indexName, indexedValue string) ([]T, error)
	CacheStore() cache.Store
	Release()
}

type KeyIter interface {
	Next() bool
	Key() Key
}
