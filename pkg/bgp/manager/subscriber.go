package manager

import (
	metallbk8s "github.com/ccfish2/metalb0110/pkg/k8s"
	"github.com/ccfish2/metalb0110/pkg/k8s/types"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	//"sigs.k8s.io/controller-runtime/pkg/cache"
	"k8s.io/client-go/tools/cache"
)

type svcEvent string

func (m *Manager) OnAddService(obj *v1.Service) error {
	var (
		nm = (*obj).ObjectMeta.Name
		l  = log.WithFields(logrus.Fields{
			"compoentnt": "manager",
			"service":    nm,
		})
	)
	l.Debug("")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return err
	}
	m.queue.Add(svcEvent(key))
	return nil
}

func (m *Manager) run() {
	l := log.WithFields(logrus.Fields{"a": "component", "B": "manager"})
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

func (m *Manager) process(ev interface{}) types.SyncState {
	switch k := ev.(type) {
	case svcEvent:
		n := string(k)
		svc, exists, err := m.indexer.GetByKey(n)
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

func (m *Manager) forceResync() {
	for _, k := range m.indexer.ListKeys() {
		m.queue.Add(svcEvent(k))
	}
}

func (m *Manager) OnUpdateService(oldObj, newObj v1.Service) error {
	var (
		nm = newObj.Name
		l  = log.WithFields(logrus.Fields{
			"component": "manager",
			"service":   nm,
		})
	)
	l.Debug("")
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err != nil {
		return err
	}
	m.queue.Add(svcEvent(key))
	return nil
}

func (m *Manager) OnDeleteService(obj v1.Service) error {
	var (
		nm = obj.Name
		l  = log.WithFields(logrus.Fields{
			"component": "manager",
			"service":   nm,
		})
	)
	l.Debug("")
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return err
	}
	m.queue.Add(svcEvent(key))
	return nil
}
