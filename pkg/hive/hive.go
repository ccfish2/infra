package hive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/dig"

	// myself
	"github.com/ccfish2/infra/pkg/hive/metrics"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "hive")
)

const (
	defaultStartTimeout = 5 * time.Minute

	defaultStopTimeout = time.Minute

	defaultEnvPrefix = "DOLPHIN_"
)

type Hive struct {
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
}

func New(cells ...cell.Cell) *Hive {
	h := &Hive{
		container:       dig.New(),
		envPrefix:       defaultEnvPrefix,
		cells:           cells,
		viper:           viper.New(),
		startTimeout:    defaultStartTimeout,
		stopTimeout:     defaultStopTimeout,
		flags:           pflag.NewFlagSet("", pflag.ContinueOnError),
		lifecycle:       &cell.DefaultLifecycle{},
		shutdown:        make(chan error, 1),
		configOverrides: nil,
	}

	if err := h.provideDefaults(); err != nil {
		log.WithError(err).Fatal("Failed to provide default objects")
	}

	if err := metrics.Cell.Apply(h.container); err != nil {
		log.WithError(err).Fatal("Failed to apply Hive metrics cell")
	}

	if err := h.container.Provide(func(healthMetrics *metrics.HealthMetrics, lc cell.Lifecycle) cell.Health {
		hp := cell.NewHealthProvider()
		updateStats := func() {
			for l, c := range hp.Stats() {
				healthMetrics.HealthStatusGauge.WithLabelValues(strings.ToLower(string(l))).Set(float64(c))
			}
		}
		lc.Append(cell.Hook{
			OnStart: func(ctx cell.HookContext) error {
				updateStats()
				hp.Subscribe(ctx, func(u cell.Update) {
					updateStats()
				}, func(err error) {})
				return nil
			},
			OnStop: func(ctx cell.HookContext) error {
				return hp.Stop(ctx)
			},
		})
		return hp
	}); err != nil {
		log.WithError(err).Fatal("Failed to provide health provider")
	}

	for _, cell := range cells {
		if err := cell.Apply(h.container); err != nil {
			log.WithError(err).Fatal("Failed to apply cell")
		}
	}

	h.flags.VisitAll(func(f *pflag.Flag) {
		if err := h.viper.BindPFlag(f.Name, f); err != nil {
			log.Fatalf("BindPFlag: %s", err)
		}
		if err := h.viper.BindEnv(f.Name, h.getEnvName(f.Name)); err != nil {
			log.Fatalf("BindEnv: %s", err)
		}
	})

	return h
}

func (h *Hive) RegisterFlags(flags *pflag.FlagSet) {
	h.flags.VisitAll(func(f *pflag.Flag) {
		if flags.Lookup(f.Name) != nil {
			log.Fatalf("Error registering flag: '%s' already registered", f.Name)
		}
		flags.AddFlag(f)
	})
}

func (h *Hive) Viper() *viper.Viper {
	return h.viper
}

type defaults struct {
	dig.Out

	Flags             *pflag.FlagSet
	Lifecycle         cell.Lifecycle
	Logger            logrus.FieldLogger
	Shutdowner        Shutdowner
	InvokerList       cell.InvokerList
	EmptyFullModuleID cell.FullModuleID
}

func (h *Hive) provideDefaults() error {
	return h.container.Provide(func() defaults {
		return defaults{
			Flags:             h.flags,
			Lifecycle:         h.lifecycle,
			Logger:            log,
			Shutdowner:        h,
			InvokerList:       h,
			EmptyFullModuleID: nil,
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

func (h *Hive) Run() error {
	startCtx, cancel := context.WithTimeout(context.Background(), h.startTimeout)
	defer cancel()

	var errs error
	if err := h.Start(startCtx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to start: %w", err))
	}

	if errs == nil {
		errs = errors.Join(errs, h.waitForSignalOrShutdown())
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), h.stopTimeout)
	defer cancel()

	if err := h.Stop(stopCtx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to stop: %w", err))
	}
	return errs
}

func (h *Hive) waitForSignalOrShutdown() error {
	signals := make(chan os.Signal, 1)
	defer signal.Stop(signals)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-signals:
		log.WithField("signal", sig).Info("Signal received")
		return nil
	case err := <-h.shutdown:
		return err
	}
}

func (h *Hive) Populate() error {
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

func (h *Hive) Start(ctx context.Context) error {
	if err := h.Populate(); err != nil {
		return err
	}

	defer close(h.fatalOnTimeout(ctx))

	log.Info("Starting")

	return h.lifecycle.Start(ctx)
}

func (h *Hive) Stop(ctx context.Context) error {
	defer close(h.fatalOnTimeout(ctx))
	log.Info("Stopping")
	return h.lifecycle.Stop(ctx)
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
			log.Fatal("Start or stop failed to finish on time, aborting forcefully.")
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

func (h *Hive) PrintObjects() {
	if err := h.Populate(); err != nil {
		log.WithError(err).Fatal("Failed to populate object graph")
	}

	fmt.Printf("Cells:\n\n")
	ip := cell.NewInfoPrinter()
	for _, c := range h.cells {
		c.Info(h.container).Print(2, ip)
		fmt.Println()
	}
	h.lifecycle.PrintHooks()
}

func (h *Hive) PrintDotGraph() {
	if err := h.Populate(); err != nil {
		log.WithError(err).Fatal("Failed to populate object graph")
	}

	if err := dig.Visualize(h.container, os.Stdout); err != nil {
		log.WithError(err).Fatal("Failed to Visualize()")
	}
}

func (h *Hive) getEnvName(option string) string {
	under := strings.Replace(option, "-", "_", -1)
	upper := strings.ToUpper(under)
	return h.envPrefix + upper
}
