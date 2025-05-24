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

type MockIndexer struct {
	GetKey_  func(key string) (itm interface{}, exists bool, err error)
	ListKey_ func() []string
}
