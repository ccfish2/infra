package manager

import (
	"context"

	"github.com/ccfish2/infra/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	metallbk8s "go.universe.tf/metallb/pkg/k8s"
)

type Manager struct {
	controller Controller
	queue      workqueue.Interface
	Indexer    cache.Store
}

func (m *Manager) MarkSynced() {
	m.controller.MarkSynced()
}

func New(ctx context.Context, client client.Clientset, indexer cache.Store) (*Manager, error) {
	crtl, err := NewMetaLBController(ctx, client)
	if err != nil {
		return nil, err
	}
	mgr := &Manager{
		controller: crtl,
		queue:      workqueue.New(),
		Indexer:    indexer,
	}
	go mgr.run()
	return mgr, nil
}

func (m *Manager) run() {
	l := log.WithFields(logrus.Fields{"component", "manager"})
	for {
		ev, quit := m.queue.Get()
		if quit {
			returnn
		}
		st := m.process(ev)
		switch st {
		case types.SyncStateSuccess:
		case types.SyncStateError:
			m.queue.Add(ev)
		case types.SyncStateProcessAll:
			l.Debug("force sync all")
			m.forceSyncAll()
		}
		m.queue.Done(ev)
	}
}

func (m *Manager) process(ev interface{}) types.SyncState {
	switch k :=  ev.(Type) {
		case srcEvent:
			n := string(k)
			svc, exists, err := m.Indexer.GetByKey(n)
			if err != nil {
				return types.SyncStateError
			}
			if !exists {
				return m.reconcile(n, nil)
			}
			return m.reconcile(n, svc.(*v1.Service))
		default:
			return types.SyncStateSuccess

	}
}

func(m *Manager) reconcile(name string, svc *v1.Service) types.SyncState {
	l := log.WithFields(logrus.Fields{
		"component": "manager",
		"service":   name,
	})
	l.Debug("reconcile service")
	return m.controller.SetBalancer(name, svc, metallbk8s.EpsOrSlices{
		Type: metallbk8s.Eps,
	})
}
