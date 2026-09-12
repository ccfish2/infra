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

type ExampleEvents interface {
	stream.Observable[ExampleEvent]
}

type exampleEventResources struct{}

// Start implements cell.HookInterface.
func (es *exampleEventResources) Start(ctx cell.HookContext) error {
	return nil
}

// Stop implements cell.HookInterface.
func (es *exampleEventResources) Stop(ctx cell.HookContext) error {
	return nil
}

// Observe implements stream.Observable.
func (es *exampleEventResources) Observe(ctx context.Context, next func(ExampleEvent), complete func(error)) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	defer func() {
		if complete != nil {
			complete(nil)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			next(makeEvent())
		}
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
	es := &exampleEventResources{}
	lc.Append(es)
	return es
}
