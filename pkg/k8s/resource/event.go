package resource

import (
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
)

const (
	Sync   EventKind = "sync"
	Upsert EventKind = "upsert"
	Delete EventKind = "delete"
)

type EventKind string

type Event[T k8sRuntime.Object] struct {
	Kind   EventKind
	Key    Key
	Object T

	Done func(err error)
}
