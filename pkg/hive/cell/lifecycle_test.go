package cell_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/stretchr/testify/assert"
)

var (
	errLifecycle = errors.New("nope")

	started, stopped int
	goodHook         = cell.Hook{
		OnStart: func(hc cell.HookContext) error {
			started++
			return nil
		},
		OnStop: func(hc cell.HookContext) error {
			stopped++
			return nil
		},
	}

	badStartHook = cell.Hook{
		OnStart: func(hc cell.HookContext) error {
			return errLifecycle
		},
	}

	badStopHook = cell.Hook{
		OnStart: func(hc cell.HookContext) error {
			started++
			return nil
		},
		OnStop: func(hc cell.HookContext) error {
			return errLifecycle
		},
	}
	nilHook = cell.Hook{nil, nil}
)

func Test_Lifecycle(t *testing.T) {
	var lc cell.DefaultLifecycle
	lc = cell.DefaultLifecycle{}
	err := lc.Start(context.TODO())
	assert.NoError(t, err, "start expected to succeed")
	err = lc.Stop(context.TODO())
	assert.NoError(t, err, "stop expected to suceed")

	// three good hook and one nil hook
	lc.Append(goodHook)
	lc.Append(goodHook)
	lc.Append(goodHook)
	lc.Append(nilHook)
	err = lc.Start(context.TODO())
	assert.NoError(t, err, "start expected to suceed")
	err = lc.Stop(context.TODO())
	assert.NoError(t, err, "stop expected to suceed")

	assert.Equal(t, started, 3)
	assert.Equal(t, stopped, 3)
	started = 0
	stopped = 0

	lc = cell.DefaultLifecycle{}
	lc.Append(goodHook)
	lc.Append(goodHook)
	lc.Append(badStartHook)
	err = lc.Start(context.TODO())
	assert.ErrorIs(t, err, errLifecycle, "start expected to fail")
	assert.Equal(t, started, 2)
	started = 0
	stopped = 0

	lc = cell.DefaultLifecycle{}
	lc.Append(goodHook)
	lc.Append(goodHook)
	lc.Append(badStopHook)
	err = lc.Start(context.TODO())
	assert.NoError(t, err, "start expected to succeed")
	assert.Equal(t, started, 3)
	assert.Equal(t, stopped, 0)

	err = lc.Stop(context.TODO())
	assert.ErrorIs(t, err, errLifecycle, "stop expected to fail")
	assert.Equal(t, stopped, 2)
	started = 0
	stopped = 0

	lc = cell.DefaultLifecycle{}
	lc.Append(cell.Hook{
		OnStop: func(hc cell.HookContext) error {
			stopped++
			return nil
		},
	})
	err = lc.Start(context.TODO())
	assert.NoError(t, err, "start expected to succeed")
	err = lc.Stop(context.TODO())
	assert.NoError(t, err, "stop expected to succeed")
	assert.Equal(t, stopped, 1)
	started = 0
	stopped = 0
}

func Test_LifecycleWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var lc cell.Lifecycle
	lc = &cell.DefaultLifecycle{}

	lc.Append(cell.Hook{
		OnStart: func(ctx cell.HookContext) error {
			<-ctx.Done()
			return ctx.Err()
		},
	})
	err := lc.Start(ctx)
	assert.ErrorIs(t, err, context.Canceled)

	lc = &cell.DefaultLifecycle{}
	expectedErr := errors.New("stop cancelled")
	ctx, cancel = context.WithCancel(context.Background())
	inStop := make(chan interface{})
	lc.Append(cell.Hook{
		OnStop: func(ctx cell.HookContext) error {
			close(inStop)
			<-ctx.Done()
			assert.ErrorIs(t, ctx.Err(), context.Canceled)
			return expectedErr
		},
	})

	go func() {
		<-inStop
		cancel()
	}()

	err = lc.Start(ctx)
	assert.NoError(t, err, "expected start to succeed")

	err = lc.Stop(ctx)
	assert.ErrorIs(t, err, expectedErr)
}
