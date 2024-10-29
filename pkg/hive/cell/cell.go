package cell

import (
	"go.uber.org/dig"
)

type Cell interface {
	Info(container) Info
}

type container interface {
	Provide(ctor any, opts ...dig.ProvideOption) error
}
