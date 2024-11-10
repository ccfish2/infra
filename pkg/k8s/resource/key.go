package resource

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"
)

type Key struct {
	Name      string
	Namespace string
}

func (k Key) String() string {
	if len(k.Namespace) > 0 {
		return k.Namespace + "/" + k.Name
	}
	return k.Name
}

func NewKey(obj any) Key {
	if d, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		namespace, name, _ := cache.SplitMetaNamespaceKey(d.Key)
		return Key{name, namespace}
	}

	meta, err := meta.Accessor(obj)
	if err != nil {
		return Key{}
	}
	if len(meta.GetNamespace()) > 0 {
		return Key{meta.GetName(), meta.GetNamespace()}
	}
	return Key{meta.GetName(), ""}
}
