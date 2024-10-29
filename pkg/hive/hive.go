package hive

import (
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
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
