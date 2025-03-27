package manager

import (
	"context"

	"github.com/ccfish2/infra/pkg/k8s/client"
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

func New(ctx context.Context, client client.Clientset, indexer cache.Store) (*Manager, error) {
	crtl, err := NewMetaLBController(ctx, client)
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
