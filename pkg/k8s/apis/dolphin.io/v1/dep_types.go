package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=false
// +kubebuilder:resource:categories={dolphin},singular="dolphinendpoint",path="dolphinendpoint",scope="Namespaced",shortName={dep}
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",description="The age of the identity",name="Age",type=date
// +kubebuilder:storageversion

// DolphinEndpoint is the Schema for the dolphinendpoints API
type DolphinEndpoint struct {
	// +deepequal-gen=false
	metav1.TypeMeta `json:",inline"`
	// +deepequal-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DolphinEndpointSpec   `json:"spec,omitempty"`
	Status DolphinEndpointStatus `json:"status,omitempty"`
}

// DolphinEndpointSpec defines the desired state of DolphinEndpoint
type DolphinEndpointSpec struct {
	Foo            string        `json:"foo,omitempty"`
	InstanceId     string        `json:"instance_id,omitempty"`
	BootID         string        `json:"boot_id,omitempty"`
	Addresses      []AddressType `json:"addresses,omitempty"`
	IngressAddress AddressPair   `json:"ingress_address,omitempty"`
	Identity       uint64        `json:"identity,omitempty"`
}

type AddressPair struct {
	IPV4 string `json:"ipv_4,omitempty"`
	IPV6 string `json:"ipv_6,omitempty"`
}

type AddressingType string

type AddressType struct {
	Addresstype AddressingType `json:"addresstype,omitempty"`
	IP          string         `json:"ip,omitempty"`
}

const (
	NodeHostName    AddressingType = "HostName"
	NodeInternalIP  AddressingType = "InternalIP"
	NodeExternalIP  AddressingType = "ExternalIP"
	NodeInternalDNS AddressingType = "InternalDNS"
	NodeExternalDNS AddressingType = "ExternalDNS"
)

type AllocationIP struct {
	Owner     string `json:"owner,omitempty"`
	Reference string `json:"reference,omitempty"`
}

// DolphinEndpointStatus defines the observed state of DolphinEndpoint
type DolphinEndpointStatus struct {
	Version          string                  `json:"version,omitempty"`
	State            map[string]string       `json:"state,omitempty"`
	AllocationStatus map[string]AllocationIP `json:"allocation_status,omitempty"`
	Checksum         int64                   `json:"checksum,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=false
// +deepequal-gen=false

// DolphinEndpointList contains a list of DolphinEndpoint
type DolphinEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DolphinEndpoint `json:"items"`
}
