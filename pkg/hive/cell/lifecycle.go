package cell

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sync"

	// dolphin
	"github.com/ccfish2/infra/pkg/hive/internal"
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
func (lc *DefaultLifecycle) Append(hook HookInterface) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.hooks = append(lc.hooks, augmentedHook{hook, nil})
}

func (lc *DefaultLifecycle) Start(ctx context.Context) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, hook := range lc.hooks {
		_, exists := getHookFuncName(hook, true)

		if !exists {
			lc.numStarted++
			continue
		}

		fmt.Printf("Executing start hook")
		t0 := time.Now()
		if err := hook.Start(ctx); err != nil {

			return err
		}
		d := time.Since(t0)
		fmt.Printf("after %d", d)
		lc.numStarted++
	}
	return nil
}

func (lc *DefaultLifecycle) Stop(ctx context.Context) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var errs error
	for ; lc.numStarted > 0; lc.numStarted-- {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		hook := lc.hooks[lc.numStarted-1]

		fnName, exists := getHookFuncName(hook, false)
		if !exists {
			continue
		}
		_ = fmt.Sprintf("function", fnName)
		fmt.Printf("Executing stop hook")
		t0 := time.Now()
		if err := hook.Stop(ctx); err != nil {
			fmt.Printf("Stop hook failed")
			errs = errors.Join(errs, err)
		} else {
			_ = time.Since(t0)
			fmt.Printf("Stop hook executed")
		}
	}
	return errs
}

func (lc *DefaultLifecycle) PrintHooks() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	fmt.Printf("Start hooks:\n\n")
	for _, hook := range lc.hooks {
		fnName, exists := getHookFuncName(hook.HookInterface, true)
		if !exists {
			continue
		}
		fmt.Printf("  • %s (%s)\n", fnName, hook.moduleID)
	}

	fmt.Printf("\nStop hooks:\n\n")
	for i := len(lc.hooks) - 1; i >= 0; i-- {
		hook := lc.hooks[i]
		fnName, exists := getHookFuncName(hook.HookInterface, false)
		if !exists {
			continue
		}
		fmt.Printf("  • %s (%s)\n", fnName, hook.moduleID)
	}
}

type Hook struct {
	OnStart func(HookContext) error
	OnStop  func(HookContext) error
}

func (h Hook) Start(ctx HookContext) error {
	if h.OnStart == nil {
		return nil
	}
	return h.OnStart(ctx)
}

func (h Hook) Stop(ctx HookContext) error {
	if h.OnStop == nil {
		return nil
	}
	return h.OnStop(ctx)
}

func getHookFuncName(hook HookInterface, start bool) (name string, hasHook bool) {
	switch hook := hook.(type) {
	case augmentedHook:
		name, hasHook = getHookFuncName(hook.HookInterface, start)
		return
	case Hook:
		if start {
			if hook.OnStart == nil {
				return "", false
			}
			return internal.FuncNameAndLocation(hook.OnStart), true
		}
		if hook.OnStop == nil {
			return "", false
		}
		return internal.FuncNameAndLocation(hook.OnStop), true

	default:
		if start {
			return internal.PrettyType(hook) + ".Start", true
		}
		return internal.PrettyType(hook) + ".Stop", true

	}
}

type augmentedLifecycle struct {
	*DefaultLifecycle
	moduleID FullModuleID
}

var _ Lifecycle = &DefaultLifecycle{}
