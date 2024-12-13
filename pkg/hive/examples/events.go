package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/stream"
	"golang.org/x/exp/rand"
)

var eventCell = cell.Provide(newExampleEvents)

type ExampleEvent struct {
	Message string
}

// Observe implements stream.Observable.
func (e ExampleEvent) Observe(ctx context.Context, next func(ExampleEvent), complete func(error)) {
	panic("unimplemented")
}

type ExampleEvents interface {
	stream.Observable[ExampleEvent]
}

type exampleEventResources struct {
	stream.Observable[ExampleEvent]

	emit     func(ExampleEvents)
	complete func(error)
}

// Start implements cell.HookInterface.
func (es exampleEventResources) Start(ctx cell.HookContext) error {
	panic("unimplemented")
}

// Stop implements cell.HookInterface.
func (es exampleEventResources) Stop(ctx cell.HookContext) error {
	panic("unimplemented")
}

func (es *exampleEventResources) emitter(ctx context.Context) error {
	timer := time.NewTicker(500 * time.Millisecond)
	select {
	case <-ctx.Done():
		return nil
	case <-timer.C:
		fmt.Println("working pool is doing something")
		return nil
	}
}

func makeEvent() ExampleEvent {
	var prefixes = []string{
		"",
		"",
		"",
	}
	prefixIndex := rand.Intn(len(prefixes))
	percent := rand.Intn(100)
	return ExampleEvent{
		Message: fmt.Sprintf("index %d percent %d", prefixIndex, percent),
	}
}

func newExampleEvents(lc cell.Lifecycle) ExampleEvents {
	es := exampleEventResources{}
	// do some business logic
	lc.Append(es)
	return es
}
