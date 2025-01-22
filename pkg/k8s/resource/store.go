package resource

import (
	corev1 "k8s.io/api/core/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// implements cache.Index
type typedStore[T k8sRuntime.Object] struct {
	store   cache.Indexer
	release func()
}

var _ Store[*corev1.Node] = &typedStore[*corev1.Node]{}

// ByIndex implements Store.
func (t *typedStore[T]) ByIndex(indexName string, indexedValue string) ([]T, error) {
	itemAnys, err := t.store.ByIndex(indexName, indexedValue)
	if err != nil {
		return nil, err
	}
	res := make([]T, len(itemAnys))
	for _, itm := range itemAnys {
		res = append(res, itm.(T))
	}
	return res, nil
}

// CacheStore implements Store.
func (t *typedStore[T]) CacheStore() cache.Store {
	return t.store
}

// Get implements Store.
func (t *typedStore[T]) Get(obj T) (item T, exists bool, err error) {
	return t.GetByKey(NewKey(obj))
}

// GetByKey implements Store.
func (t *typedStore[T]) GetByKey(key Key) (item T, exists bool, err error) {
	var itemAny any
	itemAny, exists, err = t.store.GetByKey(key.String())
	if exists {
		item = itemAny.(T)
	}
	return
}

// IndexKeys implements Store.
func (t *typedStore[T]) IndexKeys(indexName string, indexedValue string) ([]string, error) {
	return t.store.IndexKeys(indexName, indexedValue)
}

// IterKeys implements Store.
func (t *typedStore[T]) IterKeys() KeyIter {
	return &keyIterImpl{
		keys: t.store.ListKeys(),
		pos:  -1,
	}
}

type keyIterImpl struct {
	keys []string
	pos  int
}

func (it *keyIterImpl) Next() bool {
	it.pos++
	return it.pos < len(it.keys)
}

func (it *keyIterImpl) Key() Key {
	ns, nm, _ := cache.SplitMetaNamespaceKey(it.keys[it.pos])
	return Key{
		Namespace: ns,
		Name:      nm,
	}
}

// List implements Store.
func (t *typedStore[T]) List() []T {
	items := t.store.List()
	res := make([]T, len(items))
	for _, itm := range items {
		res = append(res, itm.(T))
	}
	return res
}

// Release implements Store.
func (t *typedStore[T]) Release() {
	t.release()
}

/* wrapper for cache.Store */
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
