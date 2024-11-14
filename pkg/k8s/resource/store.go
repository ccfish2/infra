package resource

import (
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type typedStore[T k8sRuntime.Object] struct {
	store   cache.Indexer
	release func()
}

// ByIndex implements Store.
func (t *typedStore[T]) ByIndex(indexName string, indexedValue string) ([]T, error) {
	panic("unimplemented")
}

// CacheStore implements Store.
func (t *typedStore[T]) CacheStore() cache.Store {
	panic("unimplemented")
}

// Get implements Store.
func (t *typedStore[T]) Get(obj T) (item T, exists bool, err error) {
	panic("unimplemented")
}

// GetByKey implements Store.
func (t *typedStore[T]) GetByKey(key Key) (item T, exists bool, err error) {
	panic("unimplemented")
}

// IndexKeys implements Store.
func (t *typedStore[T]) IndexKeys(indexName string, indexedValue string) ([]string, error) {
	panic("unimplemented")
}

// IterKeys implements Store.
func (t *typedStore[T]) IterKeys() KeyIter {
	panic("unimplemented")
}

// List implements Store.
func (t *typedStore[T]) List() []T {
	panic("unimplemented")
}

// Release implements Store.
func (t *typedStore[T]) Release() {
	panic("unimplemented")
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
