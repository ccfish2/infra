package manager

import (
	"context"
	"os"

	bgpconfig "github.com/ccfish2/infra/pkg/bgp/config"
	bgpk8s "github.com/ccfish2/infra/pkg/bgp/k8s"
	bgplog "github.com/ccfish2/infra/pkg/bgp/log"
	"github.com/ccfish2/infra/pkg/k8s/client"
	"github.com/ccfish2/infra/pkg/option"
	metallballoc "github.com/ccfish2/metalb0110/pkg/allocator"
	metallbctl "github.com/ccfish2/metalb0110/pkg/controller"
	"github.com/ccfish2/metalb0110/pkg/k8s"
	"github.com/ccfish2/metalb0110/pkg/k8s/types"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Controller interface {
	SetBalancer(name string, srvRo *v1.Service, eps k8s.EpsOrSlices) types.SyncState
	MarkSynced()
}

type metalLBController struct {
	c      *metallbctl.Controller
	logger *bgplog.Logger
}

<<<<<<< HEAD
func newMetalLBController(ctx context.Context, cs client.Clientset) (Controller, error) {
	logger := &bgplog.Logger{Entry: log}
	c := &metallbctl.Controller{
		Client: bgpk8s.New(cs, logger.Logger),
=======
func NewMetaLBController(ctx context.Context, cs client.Clientset) (Controller, error) {
	logger := &bgplog.Logger{Entry: log}
	c := &metallbctl.Controller{
		Client: bgpk8s.New(logger.Logger, cs),
>>>>>>> 4c5ccb4 (Fixing error introduced by importing)
		IPs:    metallballoc.New(),
	}

	f, err := os.Open(option.Config.BGPConfigPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config, err := bgpconfig.Parse(f)
	if err != nil {
		return nil, err
	}
<<<<<<< HEAD
	c.SetConfig(logger, config)

	return &metalLBController{
		c,
		logger,
	}, nil
=======
	c.SetConfig(logger, &cfg)
	return &metalLBController{c, logger}, nil
>>>>>>> 4c5ccb4 (Fixing error introduced by importing)
}

func (c *metalLBController) SetBalancer(name string, srvRo *v1.Service, eps k8s.EpsOrSlices) types.SyncState {
	var (
		l = log.WithFields(logrus.Fields{
			"component": "metalLBController.SetBalancer",
			"service":   name,
		})
	)
	l.Debug("assigning load balancer ip for service")
	return c.c.SetBalancer(c.logger, name, srvRo, eps)
}

func (c *metalLBController) MarkSynced() {
	c.c.MarkSynced(c.logger)
}
