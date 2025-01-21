package apis

import (
	"fmt"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io/client"
	k8sClient "github.com/ccfish2/infra/pkg/k8s/client"
)

const (
	SkipCRDCreation = "skip-crd-creation"
)

var RegisterCRDCreation = "skip-crd-creation"

var RegisterCRDsCell = cell.Module(
	"create-crds",
	"Create Dolphin CRDs",

	cell.Config(defaultConfig),

	cell.Invoke(createCRDs),

	cell.ProvidePrivate(
		newDolphinGroupCRDs,
	),
)

func (c RegisterCRDsConfig) Flags(flags *pflag.FlagSet) {
	flags.Bool(SkipCRDCreation, false, "When true, Kubernetes Custom Resource Definitions will not be created")
}

type RegisterCRDsConfig struct {
	SkipCRDCreation bool
}

var defaultConfig = RegisterCRDsConfig{}

type RegisterCRDsFunc func(k8sClient.Clientset) error

type params struct {
	cell.In

	Logger    logrus.FieldLogger
	Lifecycle cell.Lifecycle

	Clientset k8sClient.Clientset

	Config            RegisterCRDsConfig
	RegisterCRDsFuncs []RegisterCRDsFunc `group:"register-crd-funcs"`
}

func createCRDs(p params) {
	p.Lifecycle.Append(cell.Hook{
		OnStart: func(hc cell.HookContext) error {

			if !p.Clientset.IsEnabled() || p.Config.SkipCRDCreation {
				p.Logger.Info("Skipping creation of CRDs")
				return nil
			}
			p.Logger.Info("Creating CRDs ...")
			for _, f := range p.RegisterCRDsFuncs {
				if err := f(p.Clientset); err != nil {
					p.Logger.Error("Unalbe to create CRDs ", err)
					return fmt.Errorf("unable to create CRDs: %w", err)
				}
				p.Logger.Info("Complete register CRDs %v", f)
			}
			return nil
		},
	})
}

type registerCRDsFuncOut struct {
	cell.Out

	Func RegisterCRDsFunc `group:"register-crd-funcs"`
}

func newDolphinGroupCRDs() registerCRDsFuncOut {
	return registerCRDsFuncOut{
		Func: client.RegisterCRDs,
	}
}
