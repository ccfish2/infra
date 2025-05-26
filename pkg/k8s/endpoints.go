package k8s

import (
	"github.com/ccfish2/infra/pkg/k8s/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmtypes "github.com/ccfish2/infra/pkg/clustermesh/types"
	serviceStore "github.com/ccfish2/infra/pkg/service/store"
)

// +k8s:deepcopy-gen=true
// +deepequal-gen=true
type Backend struct {
	Ports         serviceStore.PortConfiguration
	NodeName      string
	Terminating   bool
	HintsForZones []string
	Preferred     bool
}

type Endpoints struct {
	types.UnserializableObject
	metav1.ObjectMeta

	EndpointSliceID

	Backends map[cmtypes.AddrCluster]*Backend
}
