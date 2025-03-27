package mock

import (
	metallbk8s "github.com/ccfish2/metalb0110/pkg/k8s"
	"github.com/ccfish2/metalb0110/pkg/k8s/types"
	v1 "k8s.io/api/core/v1"
)

type MockMetaLBController struct {
	SetBalancer_ func(name string, svc *v1.Service, eps metallbk8s.EpsOrSlices) types.SyncState
	MarkSynced_  func()
}

func (m *MockMetalLBController) SetBalancer(name string, srvRo *v1.Service, eps metallbk8s.EpsOrSlices) types.SyncState {
	return m.SetBalancer_(name, srvRo, eps)
}

func (m *MockMetalLBController) MarkSynced() {
	m.MarkSynced_()
}

// MockIndexer implements the cache.Store interface
// from the k8s.io/client-go/cache package
//
// The BGP package only utilizes two methods from this
// interface thus our mock is terse.
type MockIndexer struct {
	GetByKey_ func(key string) (item interface{}, exists bool, err error)
	ListKeys_ func() []string
}

func (m *MockIndexer) GetByKey(key string) (item interface{}, exists bool, err error) {
	return m.GetByKey_(key)
}

func (m *MockIndexer) ListKeys() []string {
	return m.ListKeys_()
}

func (m *MockIndexer) Add(obj interface{}) error {
	panic("not implemented")
}

func (m *MockIndexer) Update(obj interface{}) error {
	panic("not implemented")
}

func (m *MockIndexer) Delete(obj interface{}) error {
	panic("not implemented")
}

func (m *MockIndexer) List() []interface{} {
	panic("not implemented")
}

func (m *MockIndexer) Get(obj interface{}) (item interface{}, exists bool, err error) {
	panic("not implemented")
}

func (m *MockIndexer) Replace([]interface{}, string) error {
	panic("not implemented")
}

func (m *MockIndexer) Resync() error {
	panic("not implemented")
}
