package gops

import (
	"fmt"

	gopsAgent "github.com/google/gops/agent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/ccfish2/infra/pkg/option"
)

func Cell(defaultport uint16) cell.Cell {
	return cell.Module(
		"gops",
		"Gops Agent",
		cell.Config(GopsConfig{
			GopsPort: defaultport,
		}),
		cell.Invoke(registerGopsHook),
	)
}

type GopsConfig struct {
	GopsPort uint16
}

func (c GopsConfig) Flags(flag *pflag.FlagSet) {
	flag.Uint16(option.GopsPort, c.GopsPort, "Gops Server Listening Port")
}

func registerGopsHook(lc cell.Lifecycle, log logrus.FieldLogger, cfg GopsConfig) {
	addr := fmt.Sprintf("127.0.0.1:%d", cfg.GopsPort)

	addField := logrus.Fields{"Addr": addr, logfields.LogSubsys: "gops"}
	log.WithFields(addField)
	hk := cell.Hook{
		OnStart: func(hc cell.HookContext) error {
			if err := gopsAgent.Listen(gopsAgent.Options{Addr: addr}); err != nil {
				return err
			}
			log.Info("Gops start listening successfullyz; ")
			return nil
		},
		OnStop: func(hc cell.HookContext) error {
			gopsAgent.Close()
			log.Info("gops closes successfully")
			return nil
		},
	}
	lc.Append(hk)
	return
}
