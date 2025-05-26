package types

// +k8s:deepcopy-gen=false
// +deepequal-gen=false
type Interface interface {
	InterfaceID() string
	ForeachAddress(instanceID string, fn AddressIterator) error
	DeepCopyInterface() Interface
}

type IPAMSpec struct {
	Pool              AllocationMap `json:"pool,omitempty"`
	Pools             IPAMPoolSpec  `json:"pools,omitempty"`
	PodCIDRs          []string      `json:"podCIDRs,omitempty"`
	MinAllocate       int           `json:"minAllocate,omitempty"`
	MaxAllocate       int           `json:"maxAllocate,omitempty"`
	PreAllocate       int           `json:"preAllocate,omitempty"`
	MaxAboveWatermark int           `json:"maxAboveWatermark,omitempty"`
}

type IPAMPoolRequest struct {
	Pool   string         `json:"pool,omitempty"`
	Needed IPAMPoolDemand `json:"needed,omitempty"`
}
type IPAMPoolDemand struct {
	IPv4Addrs int `json:"ip_v4_addrs,omitempty"`
	IPv6Addrs int `json:"ip_v6_addrs,omitempty"`
}
type IPAMPodCIDR string
type IPAMPoolAllocation struct {
	Pool  string        `json:"pool,omitempty"`
	CIDRs []IPAMPodCIDR `json:"cidrs,omitempty"`
}
type IPAMPoolSpec struct {
	Requested []IPAMPoolRequest    `json:"requested,omitempty"`
	Allocated []IPAMPoolAllocation `json:"allocated,omitempty"`
}
type Address interface{}
type AddressIterator func(instanceID, interfaceID, ip, poolID string, address Address) error

type IPAMStatus struct {
	Used           AllocationMap              `json:"used,omitempty"`
	PodCIDRs       PodCIDRMap                 `json:"podCIDRs,omitempty"`
	OperatorStatus OperatorStatus             `json:"operatorStatus,omitempty"`
	ReleaseIPs     map[string]IPReleaseStatus `json:"releaseIPs,omitempty"`
}

type AllocationIP struct {
	Owner    string `json:"owner,omitempty"`
	Resource string `json:"resource,omitempty"`
}

type AllocationMap map[string]AllocationIP

type PODCIDRStaus string

const (
	PODCIDRStausReleased PODCIDRStaus = "released"
	PODCIDRStausDepleted PODCIDRStaus = "depleted"
	PODCIDRStausinUse    PODCIDRStaus = "in-use"
)

type PodCIDRMap map[string]PodCIDRMapEntry

type PodCIDRMapEntry struct {
	Status PODCIDRStaus `json:"staus,omitempty"`
}

type OperatorStatus struct {
	Error string `json:"error,omitempty"`
}

type IPReleaseStatus string
type Tags map[string]string

func (t Tags) Match(required Tags) bool {
	for k, needed := range required {
		having, ok := t[k]
		if !ok || (ok && having != needed) {
			return false
		}
	}
	return true
}
