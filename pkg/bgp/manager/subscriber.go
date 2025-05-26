package manager

import (
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

type srcEvent string

func (m *Manager) OnAddService(obj v1.Service) error {
	var (
		svcName = obj.Name
		l       = log.WithFields(logrus.Fields{
			"component":    "Manager.OnAddService",
			"service-name": svcName,
		})
	)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return err
	}
	l.Debug("adding event to queue")
	m.queue.Add(srcEvent(key))
	return nil
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

func(m *Manager) forceSyncAll() {
	for _, m := range manager.Indexer.ListKeys(){
		m.queue.Add(srcEvent(m))
	}
}

func(m *Manager) OnUpdateService(oldObj, newObj v1.Service) error {
	var (
		nm = newObj.Name
		l = log.WithFields(logrus.Fields{
			"component": "manager",
			"service": nm,
		})
	)
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err != nil {
		return err
	}
	m.queue.Add(srcEvent(key))
	return nil
}

func(m *Manager) OnDeleteService(obj v1.Service) error {
	var (
		nm = obj.Name
		l = log.WithFields(logrus.Fields{
			"component": "manager",
			"service": nm,
		})
	)
	key, err := cache.DeletionHandlingMetanamespacekeyFunnc(obj)
	if err != nil {
		return err
	}
	m.queue.Add(srcEvent(key))
	return nil
}