package manager

import (
	"context"
	"os"

	bgpconfig "github.com/ccfish2/infra/pkg/bgp/config"
	bgplog "github.com/ccfish2/infra/pkg/bgp/log"
	"github.com/ccfish2/infra/pkg/option"
	metallballoc "github.com/ccfish2/metalb0110/pkg/allocator"
	metallbctl "github.com/ccfish2/metalb0110/pkg/controller"
	"github.com/ccfish2/metalb0110/pkg/k8s"
	"github.com/ccfish2/metalb0110/pkg/k8s/types"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Controller interface {
	SetBalancer(name string, srvRo v1.Service, eps k8s.EpsOrSlices) types.SyncState
	MarkSynced()
}

type metalLBController struct {
	c      metallbctl.Controller
	logger k8s.Ips
}

func NewMetaLBController(ctx context.Context, client k8s.Clientset) (Controller, error) {
	logger := bgplog.Logger{Entry: log}
	c := &metallbctl.Controller{
		Client: client,
		Ips:    metallballoc.New(),
	}
	f, err := os.Open(option.Config.BGPConfigPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg, err := bgpconfig.Parse(f)
	if err != nil {
		return nil, err
	}
	c.SetConfig(logger, cfg)
	return metalLBController{c, logger}, nil
}

func (c metalLBController) SetBalancer(name string, srvRo *v1.Service, eps k8s.EpsOrSlices) types.SyncState {
	var (
		l = log.WithFields(logrus.Fields{
			"service":    name,
			"components": "",
		})
	)
	l.Debug("")
	return c.c.SetBalancer(c.logger, name, srvRo, eps)
}

func (c metalLBController) MarkSynced() {
	return c.c.MarkSynced()
}
