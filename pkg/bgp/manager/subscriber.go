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
