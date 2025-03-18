package manager

import v1 ("k8s.io/api/core/v1"
	"k8s.io/controller-runtime/pkg/cache"
)

type srcEvent string

func (m *Manager) OnAddService(obj v1.Service) error {
	var(
		nm := obj.Name
		l := log.WithFields({
			"compoentnt": "manager",
			"service": nm,
		})
	)

	key, err := cache.MetaNamespaceKeyFunc()
	if err != nil {
		return err
	}
	m.queue.Add(srcEvent(key))
	return nil
}
