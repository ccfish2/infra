package cell

import (
	"context"
	"errors"
	"fmt"
	"io"
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

type HookDescriptiveInterface interface {
	HookInterface
	HookInfo() string
}

type Lifecycle interface {
	Append(HookInterface)

	Start(context.Context) error
	Stop(context.Context) error
	PrintHooks(io.Writer)
}

type augmentedHook struct {
	HookInterface
	moduleID FullModuleID
}

type DefaultLifecycle struct {
	mu         sync.Mutex
	hooks      []augmentedHook
	numStarted int

	LogThreshold time.Duration
}

func (lc *DefaultLifecycle) Append(hook HookInterface) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	fmt.Printf("added hook %v \n", internal.FuncNameAndLocation(hook.Start))
	lc.hooks = append(lc.hooks, augmentedHook{hook, nil})
}

func (lc *DefaultLifecycle) Start(ctx context.Context) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fmt.Println("Lifecycle start with ", len(lc.hooks), " hooks ")
	for _, hook := range lc.hooks {
		fnName, exists := getHookFuncName(hook, true)

		if !exists {
			fmt.Println("hook ", fnName, " does not exist")
			lc.numStarted++
			continue
		}

		fmt.Println("Executing start hook ", fnName)
		t0 := time.Now()
		if err := hook.Start(ctx); err != nil {
			fmt.Println("start hook failed ", err)
			return err
		}
		d := time.Since(t0)
		fmt.Println("hook get executed for ", d)
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

		fmt.Println("Executing stop hook ", fnName)
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

type PreStopHook interface {
	PreStopHookMarker()
}

func (lc *DefaultLifecycle) PrintHooks(w io.Writer) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	fmt.Fprintf(w, "Start hooks:\n\n")
	for _, hook := range lc.hooks {
		fnName, exists := getHookFuncName(hook.HookInterface, true)
		if !exists {
			continue
		}
		fmt.Printf("  • %s (%s)\n", fnName, hook.moduleID)
	}

	fmt.Fprintf(w, "\nStop hooks:\n\n")
	printHook := func(hook augmentedHook) {
		fnName, exists := getHookFuncName(hook.HookInterface, false)
		if !exists {
			return
		}
		fmt.Fprintf(w, "  • %s (%s)\n", fnName, hook.moduleID)
	}
	for i := len(lc.hooks) - 1; i >= 0; i-- {
		hook := lc.hooks[i]
		if _, ok := hook.HookInterface.(PreStopHook); ok {
			printHook(hook)
		}
	}
	for i := len(lc.hooks) - 1; i >= 0; i-- {
		hook := lc.hooks[i]
		if _, ok := hook.HookInterface.(PreStopHook); !ok {
			printHook(hook)
		}
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
