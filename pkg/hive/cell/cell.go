package cell

import (
	"go.uber.org/dig"
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
	Scope(name string, opts ...dig.ScopeOption) error
}
