package resource

import (
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
)

type EventKind string

type Event[T k8sRuntime.Object] struct {
	Kind   EventKind
	Key    Key
	Object T

	Done func(err error)
}
