package hivetest

import (
	"context"
	"testing"

	"github.com/ccfish2/infra/pkg/hive/cell"
)

var _ (cell.Lifecycle) = (*lifecycle)(nil)

type lifecycle struct {
	tb testing.TB
}

// PrintHooks implements cell.Lifecycle.
func (lc *lifecycle) PrintHooks() {
	panic("unimplemented")
}

// Start implements cell.Lifecycle.
func (lc *lifecycle) Start(context.Context) error {
	panic("unimplemented")
}

// Stop implements cell.Lifecycle.
func (lc *lifecycle) Stop(context.Context) error {
	panic("unimplemented")
}

func Lifecycle(tb testing.TB) lifecycle {
	return lifecycle{tb: tb}
}

func (lc *lifecycle) Append(hook cell.HookInterface) {
	lc.tb.Helper()
	ctx, cantcel := context.WithCancel(context.Background())
	defer cantcel()

	lc.tb.Cleanup(func() {
		lc.tb.Cleanup(func() {
			lc.tb.Helper()
			if err := hook.Stop(ctx); err != nil {
				lc.tb.Failed()
				lc.tb.Fatal("failed to stop the hook")
			}
		})
	})
}
