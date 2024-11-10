package resource

import (
	"github.com/ccfish2/infra/pkg/stream"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
)

type Resource[T k8sRuntime.Object] interface {
	stream.Observable[Event[T]]
}
