package manager

import (
	"context"

	"github.com/ccfish2/infra/pkg/k8s/client"
	metallbk8s "github.com/ccfish2/metalb0110/pkg/k8s"
	"github.com/ccfish2/metalb0110/pkg/k8s/types"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Manager struct {
	controller Controller
	queue      workqueue.Interface
	indexer    cache.Store
}

func (m *Manager) MarkSynced() {
	m.controller.MarkSynced()
}

type svcEvent string

func New(ctx context.Context, client client.Clientset, indexer cache.Store) (*Manager, error) {
	crtl, err := newMetalLBController(ctx, client)
	if err != nil {
		return nil, err
	}
	mgr := &Manager{
		controller: crtl,
		queue:      workqueue.New(),
		indexer:    indexer,
	}
	go mgr.run()
	return mgr, nil
}

func (m *Manager) run() {
	l := log.WithFields(logrus.Fields{
		"component": "Manager.run",
	})
	for {
		ev, quit := m.queue.Get()
		if quit {
			return
		}
		st := m.process(ev)
		switch st {
		case types.SyncStateSuccess:
		case types.SyncStateError:
			m.queue.Add(ev)
		case types.SyncStateReprocessAll:
			l.Debug("force sync all")
			m.forceResync()
		}
		m.queue.Done(ev)
	}
}

func (m *Manager) process(event interface{}) types.SyncState {
	switch k := event.(type) {
	case svcEvent:
		n := string(k) // service namespace/name

		svc, exists, err := m.indexer.GetByKey(n)
		if err != nil {
			return types.SyncStateError
		}
		if !exists {
			return m.reconcile(n, nil) // Causes MetalLB to unassign the LB IP
		}
		return m.reconcile(n, svc.(*v1.Service))
	default:
		log.Debugf("Encountered an unknown key type %T in BGP controller", k)
		return types.SyncStateSuccess
	}
}

func (m *Manager) forceResync() {
	for _, k := range m.indexer.ListKeys() {
		m.queue.Add(svcEvent(k))
	}
}

func (m *Manager) reconcile(name string, svc *v1.Service) types.SyncState {
	l := log.WithFields(logrus.Fields{
		"component": "manager",
		"service":   name,
	})
	l.Debug("reconcile service")
	return m.controller.SetBalancer(name, svc, metallbk8s.EpsOrSlices{
		Type: metallbk8s.Eps,
	})
}
