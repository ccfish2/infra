package hive

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

const (
	defaultStartTimeout = 5 * time.Minute
	defaultStopTimeout  = time.Minute
	defaultEnvPrefix    = "DOLPHIN_"
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

func (h *Hive) Viper() *viper.Viper {
	return h.viper
}

type defaults struct {
	dig.Out

	Flags             *pflag.FlagSet
	Lifecycle         cell.Lifecycle
	Logger            logrus.FieldLogger
	Shutdowner        Shutdowner
	InvokeList        cell.InvokerList
	EmptyFullModuleID cell.FullModuleID
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

	// create module scoped health reporters

	// Apply all cells into the containers

	// Pass all parameters to the viper

	return h
}

/* list all objects within the container*/
func (h *Hive) PrintObjects() {
	if err := h.Populate(); err != nil {
		fmt.Println("Failed to populate object graph")
	}
	fmt.Printf("Cells:\n\n")
	ip := cell.NewInfoPrinter()
	for _, c := range h.cells {
		c.Info(h.container).Print(2, ip)
		fmt.Println()
	}
	h.lifecycle.PrintHooks()
}

func (h *Hive) RegisterFlags(flags *pflag.FlagSet) {
	h.flags.VisitAll(func(f *pflag.Flag) {
		if flags.Lookup(f.Name) != nil {
			log.Fatalf("Error registering flag: '%s' already registered", f.Name)
		}
		flags.AddFlag(f)
	})
}

func (h *Hive) PrintDotGraph() {
	if err := h.Populate(); err != nil {
		fmt.Println("Failed to populate object graph")
	}
	if err := dig.Visualize(h.container, os.Stdout); err != nil {
		fmt.Println("Failed to Visualize()")
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

	for _, o := range h.configOverrides {
		v := reflect.ValueOf(o)

		t := v.Type()
		if t.Kind() != reflect.Func || t.NumIn() != 1 {
			return fmt.Errorf("config override has invalid type %T, expected func(*T)", o)
		}
		flaggerType := reflect.TypeOf((*cell.Flagger)(nil)).Elem()
		if !t.In(0).Implements(flaggerType) {
			return fmt.Errorf("config override function parameters (%T) does not implement Flagger", o)
		}

		provideFunc := func(in []reflect.Value) []reflect.Value {
			return []reflect.Value{v}
		}
		providerFuncType := reflect.FuncOf(nil, []reflect.Type{t}, false)
		pfv := reflect.MakeFunc(providerFuncType, provideFunc)
		if err := h.container.Provide(pfv.Interface()); err != nil {
			return fmt.Errorf("providing ocnfig override failed: %w", err)
		}
	}

	// Execute the invoke functions to construct the objects
	for _, invoke := range h.invokes {
		if err := invoke(); err != nil {
			return err
		}
	}
	return nil
}

func (h *Hive) Start(ctx context.Context) error {
	if err := h.Populate(); err != nil {
		return err
	}
	defer close(h.fatalOnTimeout(ctx))

	fmt.Println("Starting")

	return h.lifecycle.Start(ctx)
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
	stopContext, cancel := context.WithTimeout(context.Background(), h.stopTimeout)
	defer cancel()
	if err := h.Stop(stopContext); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to stop :%w"), err)
	}
	return errs
}

func (h *Hive) fatalOnTimeout(ctx context.Context) chan struct{} {
	// buffered, receiver consumes when it is sent
	terminated := make(chan struct{}, 1)
	go func() {
		select {
		case <-terminated:
		case <-ctx.Done():
		}

		select {
		case <-terminated:
		case <-time.After(5 * time.Second):
			fmt.Println("Start or stop failed to finish on time, aborting forcefully.")
		}
	}()
	return terminated
}

func (h *Hive) waitForSignalOrShutdown() error {
	signals := make(chan os.Signal, 1)
	defer signal.Stop(signals)
	signal.Notify(signals, os.Interrupt)
	select {
	case sig := <-signals:
		fmt.Printf("signal %v signal received", sig)
		return nil
	case err := <-h.shutdown:
		return err
	}
}

func (h *Hive) Stop(ctx context.Context) error {
	defer close(h.fatalOnTimeout(ctx))
	fmt.Println("Stopping")
	return h.lifecycle.Stop(ctx)
}
