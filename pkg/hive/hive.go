package hive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/hive/script"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

const (
	defaultStartTimeout = 5 * time.Minute

	defaultStopTimeout = time.Minute

	// default envprefix is the default prefix for environment variables
	// e.g flag "foo" can be set with environment variable "DOLPHIN_"
	defaultEnvPrefix    = "DOLPHIN_"
	defaultLogThreshold = time.Duration(0)
)

type Options struct {
	// EnvPrefix is the prefix to use for environment variables, e.g.
	EnvPrefix string

	// ModuleDecorator is an optional set of decorator functions to use for each
	// module. This can be used to provide module-scoped overrides to objects.
	ModuleDecorators cell.ModuleDecorators

	// ModulePrivateProvider is an optional set of private provide functions to
	// use for each module. This can be used to provide module-scoped objects.
	ModulePrivateProviders cell.ModulePrivateProviders

	// DecodeHooks are optional additional decode hooks to use with cell.Config
	DecodeHooks cell.DecodeHooks

	StartTimeout time.Duration
	StopTimeout  time.Duration

	LogThreshold time.Duration

	Lifecycle cell.Lifecycle
}

func DefaultOptions() Options {
	return Options{
		EnvPrefix:        defaultEnvPrefix,
		ModuleDecorators: nil,
		StartTimeout:     defaultStartTimeout,
		StopTimeout:      defaultStopTimeout,
		LogThreshold:     defaultLogThreshold,
	}
}

// framework used to build modular applications
// it implements dependency injection using the ubber dig library

type Hive struct {
	opts                      Options
	container                 *dig.Container
	cells                     []cell.Cell
	shutdown                  chan error
	envPrefix                 string
	startTimeout, stopTimeout time.Duration
	flags                     *pflag.FlagSet
	viper                     *viper.Viper
	lifecycle                 cell.Lifecycle
	populated                 bool
	invokes                   []func() error
	configOverrides           []any
	metrics                   Metrics
}

// reutrns a new Hive that can be run or injected
func New(cells ...cell.Cell) *Hive {
	return NewWithOptions(DefaultOptions(), cells...)
}

func NewWithOptions(opts Options, cells ...cell.Cell) *Hive {
	h := &Hive{
		opts:            opts,
		container:       dig.New(dig.DeferAcyclicVerification()),
		cells:           cells,
		viper:           viper.New(),
		flags:           pflag.NewFlagSet("", pflag.ContinueOnError),
		lifecycle:       opts.Lifecycle,
		metrics:         NopMetrics{},
		shutdown:        make(chan error, 1),
		configOverrides: nil,
		envPrefix:       opts.EnvPrefix,
		startTimeout:    opts.StartTimeout,
		stopTimeout:     opts.StopTimeout,
	}
	if h.lifecycle == nil {
		h.lifecycle = &cell.DefaultLifecycle{
			LogThreshold: opts.LogThreshold,
		}
	}

	if err := h.container.Provide(func(lc cell.Lifecycle) cell.Health {
		hp := cell.NewHealthProvider()
		lc.Append(cell.Hook{
			OnStop: func(ctx cell.HookContext) error {
				return hp.Stop(ctx)
			},
		})
		return hp
	}); err != nil {
		panic(fmt.Sprintf("Failed to provide health provider: %s", err))
	}

	if err := h.provideDefaults(); err != nil {
		panic(fmt.Sprintf("Failed to provide defaults: %s", err))
	}

	// Apply all cells to the container. This registers all constructors
	// and adds all config flags. Invokes are delayed until Start() is
	// called.
	for _, cell := range cells {
		if err := cell.Apply(h.container, h.container); err != nil {
			panic(fmt.Sprintf("Failed to apply cell: %s", err))
		}
	}

	// Bind the newly registered flags to viper.
	h.flags.VisitAll(func(f *pflag.Flag) {
		if err := h.viper.BindPFlag(f.Name, f); err != nil {
			panic(fmt.Sprintf("BindPFlag: %s", err))
		}
		if err := h.viper.BindEnv(f.Name, h.getEnvName(f.Name)); err != nil {
			panic(fmt.Sprintf("BindEnv: %s", err))
		}
	})

	return h
}

// adds all flags within the hive to the given flag set
func (h *Hive) RegisterFlags(flags *pflag.FlagSet) {
	h.flags.VisitAll(func(f *pflag.Flag) {
		if flags.Lookup(f.Name) != nil {
			panic(fmt.Sprintf("Error registering flag: '%s' already registered", f.Name))
		}
		flags.AddFlag(f)
	})
}

func (h *Hive) Viper() *viper.Viper {
	return h.viper
}

type defaults struct {
	dig.Out

	Flags                  *pflag.FlagSet
	Lifecycle              cell.Lifecycle
	Logger                 logrus.FieldLogger
	Shutdowner             Shutdowner
	InvokerList            cell.InvokerList
	EmptyFullModuleID      cell.FullModuleID
	DecodeHooks            cell.DecodeHooks
	ModuleDecorators       cell.ModuleDecorators
	ModulePrivateProviders cell.ModulePrivateProviders
}

func (h *Hive) provideDefaults() error {
	return h.container.Provide(func() defaults {
		return defaults{
			Flags:                  h.flags,
			Lifecycle:              h.lifecycle,
			Shutdowner:             h,
			InvokerList:            h,
			EmptyFullModuleID:      nil,
			DecodeHooks:            h.opts.DecodeHooks,
			ModuleDecorators:       h.opts.ModuleDecorators,
			ModulePrivateProviders: h.opts.ModulePrivateProviders,
			Logger:                 logging.DefaultLogger,
		}
	})
}

func (h *Hive) SetTimeouts(start, stop time.Duration) {
	h.startTimeout, h.stopTimeout = start, stop
}

func (h *Hive) SetEnvPrefix(prefix string) {
	h.envPrefix = prefix
}

func AddConfigOverride[Cfg cell.Flagger](h *Hive, override func(*Cfg)) {
	h.configOverrides = append(h.configOverrides, override)
}

type RunOptionFunc func(*Hive)

func WithStartTimeout(timeout time.Duration) RunOptionFunc {
	return func(h *Hive) {
		h.opts.StartTimeout = timeout
	}
}

func WithStopTimeout(timeout time.Duration) RunOptionFunc {
	return func(h *Hive) {
		h.opts.StopTimeout = timeout
	}
}

func WithLogThreshold(threshold time.Duration) RunOptionFunc {
	return func(h *Hive) {
		h.opts.LogThreshold = threshold
	}
}

// populates the cell configuration and runs the hive cells
// Intterupt signal or call to the Shutdowner.Shutdown() will cause the hive to stop
func (h *Hive) Run(args ...any) error {
	log, _ := parseRunArgs(args)
	startCtx, cancel := context.WithTimeout(context.Background(), h.startTimeout)
	defer cancel()

	var errs error
	if err := h.Start(log, startCtx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to start: %w", err))
	}

	if errs == nil {
		errs = errors.Join(errs, h.waitForSignalOrShutdown(log))
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), h.stopTimeout)
	defer cancel()

	if err := h.Stop(log, stopCtx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to stop: %w", err))
	}
	return errs
}

func parseRunArgs(args []any) (*slog.Logger, []RunOptionFunc) {
	var log *slog.Logger
	var opts []RunOptionFunc
	for _, arg := range args {
		switch arg := arg.(type) {
		case *slog.Logger:
			log = arg
		case RunOptionFunc:
			opts = append(opts, arg)
		}
	}
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return log, opts
}

func (h *Hive) waitForSignalOrShutdown(log *slog.Logger) error {
	signals := make(chan os.Signal, 1)
	defer signal.Stop(signals)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-signals:
		log.Info("Signal received", "signal", sig)
		return nil
	case err := <-h.shutdown:
		return err
	}
}

func (h *Hive) Populate(logs ...*slog.Logger) error {
	var log *slog.Logger
	if len(logs) > 0 {
		log = logs[0]
	}
	if log == nil {
		log, _ = parseRunArgs(nil)
	}
	if h.populated {
		return nil
	}
	h.populated = true

	err := h.container.Provide(
		func() cell.AllSettings {
			return cell.AllSettings(h.viper.AllSettings())
		})
	if err != nil {
		return err
	}

	// Provide config overriders if any
	for _, o := range h.configOverrides {
		v := reflect.ValueOf(o)
		// Check that the config override is of type func(*cfg) and
		// 'cfg' implements Flagger.
		t := v.Type()
		if t.Kind() != reflect.Func || t.NumIn() != 1 {
			return fmt.Errorf("config override has invalid type %T, expected func(*T)", o)
		}
		flaggerType := reflect.TypeOf((*cell.Flagger)(nil)).Elem()
		if !t.In(0).Implements(flaggerType) {
			return fmt.Errorf("config override function parameter (%T) does not implement Flagger", o)
		}

		providerFunc := func(in []reflect.Value) []reflect.Value {
			return []reflect.Value{v}
		}
		providerFuncType := reflect.FuncOf(nil, []reflect.Type{t}, false)
		pfv := reflect.MakeFunc(providerFuncType, providerFunc)
		if err := h.container.Provide(pfv.Interface()); err != nil {
			return fmt.Errorf("providing config override failed: %w", err)
		}
	}

	for _, invoke := range h.invokes {
		if err := invoke(); err != nil {
			return err
		}
	}
	return nil
}

func (h *Hive) AppendInvoke(invoke func() error) {
	h.invokes = append(h.invokes, invoke)
}

func (h *Hive) Start(args ...any) error {
	log, ctx := parseLifecycleArgs(args)
	if err := h.Populate(log); err != nil {
		return err
	}

	defer close(h.fatalOnTimeout(ctx))

	log.Info("Starting")

	return h.lifecycle.Start(log, ctx)
}

func (h *Hive) Stop(args ...any) error {
	log, ctx := parseLifecycleArgs(args)
	defer close(h.fatalOnTimeout(ctx))
	log.Info("Stopping")
	return h.lifecycle.Stop(log, ctx)
}

func parseLifecycleArgs(args []any) (*slog.Logger, context.Context) {
	var log *slog.Logger
	var ctx context.Context
	for _, arg := range args {
		switch arg := arg.(type) {
		case *slog.Logger:
			log = arg
		case context.Context:
			ctx = arg
		}
	}
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return log, ctx
}

func (h *Hive) fatalOnTimeout(ctx context.Context) chan struct{} {
	terminated := make(chan struct{}, 1)
	go func() {
		select {
		case <-terminated:
			return

		case <-ctx.Done():
		}

		select {
		case <-terminated:
		case <-time.After(5 * time.Second):
			panic("Start or stop failed to finish on time, aborting forcefully.")
		}
	}()
	return terminated
}

func (h *Hive) Shutdown(opts ...ShutdownOption) {
	var o shutdownOptions
	for _, opt := range opts {
		opt.apply(&o)
	}

	select {
	case h.shutdown <- o.err:
	default:
	}
}

type ErrPopulate struct {
	err error
}

// Error implements error.
func (e ErrPopulate) Error() string {
	return fmt.Sprintf("failed to populate object graph: %s", dig.RootCause(e.err))
}

func (e ErrPopulate) Pretty() string {
	return fmt.Sprintf("Failed to populate object graph:\n%+v", e.err)
}

var _ error = ErrPopulate{}

func (h *Hive) PrintObjects(w io.Writer, log *slog.Logger) error {
	if err := h.Populate(log); err != nil {
		err := ErrPopulate{err}
		fmt.Fprintf(w, "ERROR: %s\n", err.Pretty())
	}

	fmt.Printf("Cells:\n\n")
	ip := cell.NewInfoPrinter()
	for _, c := range h.cells {
		c.Info(h.container).Print(2, ip)
		fmt.Println()
	}
	h.lifecycle.PrintHooks(w)

	return nil
}

func (h *Hive) PrintDotGraph() error {
	if err := h.Populate(nil); err != nil {
		err := ErrPopulate{err}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Pretty())
		return err
	}

	return dig.Visualize(h.container, os.Stdout)
}

func (h *Hive) getEnvName(option string) string {
	under := strings.Replace(option, "-", "_", -1)
	upper := strings.ToUpper(under)
	return h.envPrefix + upper
}

func (h *Hive) ScriptCommands(log *slog.Logger) (map[string]script.Cmd, error) {
	if err := h.Populate(log); err != nil {
		return nil, fmt.Errorf("failed to populate object graph: %s", err)
	}
	m := map[string]script.Cmd{}
	m["hive"] = hiveScriptCmd(h, log)
	m["hive/start"] = hiveStartCmd(h, log)
	m["hive/stop"] = hiveStopCmd(h, log)

	// Gather the commands from the hive.
	err := h.container.Invoke(func(sc ScriptCmds) {
		for name, cmd := range sc.Map() {
			m[name] = cmd
		}
	})
	return m, err
}

func (h *Hive) provideHealthProvider() cell.Health {
	hp := cell.NewHealthProvider()
	h.lifecycle.Append(cell.Hook{
		OnStop: func(ctx cell.HookContext) error {
			return hp.Stop(ctx)
		},
	})
	return hp
}
