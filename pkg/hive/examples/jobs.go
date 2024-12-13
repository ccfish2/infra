package main

import (
	"context"
	"runtime/pprof"
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/hive/job"
	"github.com/ccfish2/infra/pkg/stream"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
	"k8s.io/client-go/util/workqueue"
)

var jobCell = cell.Provide(newExampleCell)

type exampleJobCell struct {
	jobGroup job.Group
	workChan chan struct{}
	trigger  job.Trigger
	logger   logrus.FieldLogger
}

func newExampleCell(lc cell.Lifecycle, logger logrus.FieldLogger, registry job.Registry, scope cell.Scope) *exampleJobCell {
	ex := exampleJobCell{
		jobGroup: registry.NewGroup(
			scope,
			job.WithLogger(logger),
			job.WithPprofLabels(pprof.LabelSet{}),
		),
		workChan: make(chan struct{}, 3),
		trigger:  job.NewTrigger(),
		logger:   logger,
	}
	ex.jobGroup.Add(
		job.OneShot("sync-on-all", ex.sync, job.WithRetry(3, workqueue.DefaultControllerRateLimiter()), job.WithShutdown()),
		job.OneShot("daemon", ex.daemon),
		job.Timer("timer", ex.timer, 5*time.Second, job.WithTrigger(ex.trigger)),
		job.Observer("observer", ex.Observe, stream.FromChannel(ex.workChan)),
	)
	lc.Append(ex.jobGroup)
	return &ex
}

func (ex *exampleJobCell) sync(ctx context.Context, health cell.HealthReporter) error {
	panic("")
}

func (ex *exampleJobCell) doSomework() error {
	ex.workChan = make(chan struct{})
	return nil
}

func (ex *exampleJobCell) daemon(ctx context.Context, reporter cell.HealthReporter) error {
	for {
		randTimeout := time.NewTimer(time.Duration(rand.Intn(3000)) * time.Millisecond)
		select {
		case <-ctx.Done():
			return nil
		case <-randTimeout.C:
			ex.doSomework()
			return nil
		}
	}

}

func (ex *exampleJobCell) timer(ctx context.Context) error {
	if err := ex.doSomework(); err != nil {
		return nil
	}
	return nil
}

func (ex *exampleJobCell) Trigger() {
	ex.trigger.Trigger()
}

func (ex *exampleJobCell) Observe(ctx context.Context, event struct{}) error {
	ex.logger.Info("observed event")
	return nil
}
