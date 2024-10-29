package cell

import (
	"context"

	"sync"
)

type HookContext context.Context

type HookInterface interface {
	Start(HookContext) error
	Stop(HookContext) error
}

type Lifecycle interface {
	Append(HookInterface)

	Start(context.Context) error
	Stop(context.Context) error
	PrintHooks()
}

type augmentedHook struct {
	HookInterface
	moduleID FullModuleID
}

type DefaultLifecycle struct {
	mu         sync.Mutex
	hooks      []augmentedHook
	numStarted int
}

// Append implements Lifecycle.
func (d *DefaultLifecycle) Append(HookInterface) {
	panic("unimplemented")
}

// PrintHooks implements Lifecycle.
func (d *DefaultLifecycle) PrintHooks() {
	panic("unimplemented")
}

// Start implements Lifecycle.
func (d *DefaultLifecycle) Start(context.Context) error {
	panic("unimplemented")
}

// Stop implements Lifecycle.
func (d *DefaultLifecycle) Stop(context.Context) error {
	panic("unimplemented")
}

var _ Lifecycle = &DefaultLifecycle{}
