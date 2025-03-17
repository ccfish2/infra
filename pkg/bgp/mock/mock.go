package mock

import (
	"go.universe.tf/metallb/pkg/k8s/types"
	metallbspr "go.universe.tf/metallb/pkg/speaker"
	v1 "k8s.io/api/core/v1"
)

type MockMetaLBController struct {
	SetBalancer_ func(name string, svc *v1.Service, eps metallbspr.EpsOrSlices) types.SyncState
	MarkSynced_  func()
}

type MockIndexer struct {
	GetKey_  func(key string) (itm interface{}, exists bool, err error)
	ListKey_ func() []string
}
