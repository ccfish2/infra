package cell

import (
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"go.uber.org/dig"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "hive")
)

type Cell interface {
	Info(container) Info
	Apply(container) error
}

type In = dig.In
type Out = dig.Out

// it is the object including one directed acyclic graph
type container interface {
	Provide(ctor any, opts ...dig.ProvideOption) error
	Invoke(fn any, opts ...dig.InvokeOption) error
	Decorate(fn any, opts ...dig.DecorateOption) error
	Scope(name string, opts ...dig.ScopeOption) *dig.Scope
}
