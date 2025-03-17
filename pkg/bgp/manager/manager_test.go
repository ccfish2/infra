package manager

import (
	"errors"
	"time"
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
	panic(" not implemented")
}
func TestManagerEvent(t *testing.T) {
	svc, _, svc1d, svc2d := mock.GenTestServiceParis()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)

	var rr struct{
		lock.Mutex
		name string 
		svcRo *v1.Service
		eps metallbK8s.EpsOrSlices
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
		}	
	}


	mockIndexer := &mock.MockIndexer{
		GetKey_: func(key string) (itm interface{}, exists bool, err error) {
			return &service, true, nil
		}
	}

	mgr := Manager{
		controller: mockCtrl,
		queue:      workqueue.New(),
		Indexer:    mockIndexer,
	}

	go mgr.Run()
	err := mgr.OnAddService(&service)
	if err != nil {
		panic("failed on Adding Service")
	}
	<ctx.Done()

}
