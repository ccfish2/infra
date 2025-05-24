package manager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ccfish2/infra/pkg/lock"
	"github.com/ccfish2/metalb0110/pkg/k8s/types"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
)

const (
	DefaultTimeout = 30 * time.Second
)

var (
	errTimeout = errors.New("timeout before mock received event")
	emptEps    = metallbK8s.EpsOrSlices{
		Type: metallbK8s.Eps,
	}
)

func TestManagerNoService(t *testing.T) {
	svc, __, __, svc2d := mock.GenTestServiceParis()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	var rr struct {
		lock.Mutex
		name  string
		srvRo *v1.Service
		eps   mlbk8s.EpsOrSlices
	}

	mockCtrl := &mock.MockMetalLBController{
		SetBalancer_: func(name string, srvRo *v1.Service, eps mlbk8s.EpsOrSlices) types.SyncState {
			rr.Lock()
			rr.name, rr.srvRo, rr.eps = name, srvRo, eps
			rr.Unlock()
			cancel()
			return types.SyncStateSuccess
		},
	}

	mockIndexer := &mock.MockIndexer{
		GetKey_: func(key string) (itm interface{}, exists bool, err error) {
			return &service, true, nil
		},
	}

	mgr := Manager{
		controller: mockCtrl,
		queue:      workqueue.New(),
		indexer:    mockIndexer,
	}

	go mgr.Run()
	err := mgr.OnAddService(&service)
	if err != nil {
		panic("failed on Adding Service")
	}
	<-ctx.Done()

	rr.Lock()
	defer rr.Unlock()

	if !cmp.Equal(rr.name, svc1d) {
		t.Fatal(cmp.Diff(rr.name, svc1d))
	}
	if rr.srvRo != nil {
		t.Fatal("expect nil service")
	}
	if !cmp.Equal(rr.eps, emptEps) {
		t.Fatal(cmp.Diff(rr.eps, emptEps))
	}
}

func TestManagerEvent(t *testing.T) {
	svc, _, svc1d, svc2d := mock.GenTestServiceParis()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)

	var rr struct {
		lock.Mutex
		name  string
		svcRo *v1.Service
		eps   metallbK8s.EpsOrSlices
	}

	mockCtrl := &mock.MockMetaLBController{
		SetBalancer_: func(name string, svc *v1.Service, eps metallbK8s.EpsOrSlices) types.SyncState {
			rr.Lock()

			rr.name = name
			rr.svcRo = svc
			rr.eps = eps
			rr.Unlock()
			cancel()
			return types.SyncStateSuccess
		},
	}

	mockIndexer := &mock.MockIndexer{
		GetKey_: func(key string) (itm interface{}, exists bool, err error) {
			return &service, true, nil
		},
	}

	mgr := Manager{
		controller: mockCtrl,
		queue:      workqueue.New(),
		indexer:    mockIndexer,
	}

	go mgr.Run()
	err := mgr.OnAddService(&service)
	if err != nil {
		panic("failed on Adding Service")
	}
	<-ctx.Done()

	if !cmp.Equal(rr.name, svc1d) {
		t.Fatal(cmp.Diff(rr.name, svc1d))
	}
	if !cmp.Equal(rr.svcRo, &service) {
		t.Fatal(cmp.Diff(rr.svcRo, &service))
	}
	if !cmp.Equal(rr.eps, emptEps) {
		t.Fatal(cmp.Diff(rr.eps, emptEps))
	}
}
