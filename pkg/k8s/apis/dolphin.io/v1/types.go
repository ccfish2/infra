package v1

import (
	"net"
	"sort"

	eniTypes "github.com/ccfish2/infra/pkg/aws/eni/types"
	azureTypes "github.com/ccfish2/infra/pkg/azure/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// myself
	ipamTypes "github.com/ccfish2/infra/pkg/ipam/types"
	"github.com/ccfish2/infra/pkg/node/addressing"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={dolphin},singular="dolphinidentity",path="dolphinidentities",scope="Cluster",shortName={dolphinid}
// +kubebuilder:printcolumn:JSONPath=".metadata.labels.io\\.kubernetes\\.pod\\.namespace",description="The namespace of the entity",name="Namespace",type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",description="The age of the identity",name="Age",type=date
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// DolphinIdentity is a CRD that represents an identity managed by Dolphin.
//
//	kubectl get dolphinid -l 'foo=bar'
type DolphinIdentity struct {
	// +deepequal-gen=false
	metav1.TypeMeta `json:",inline"`
	// +deepequal-gen=false
	metav1.ObjectMeta `json:"metadata"`

	// SecurityLabels is the source-of-truth set of labels for this identity.
	SecurityLabels map[string]string `json:"security-labels"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +deepequal-gen=false

// DolphinIdentityList is a list of DolphinIdentity objects.
type DolphinIdentityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of DolphinIdentity
	Items []DolphinIdentity `json:"items"`
}

// +k8s:deepcopy-gen=false

// AddressPairList is a list of address pairs.
type AddressPairList []*AddressPair

// Sort sorts an AddressPairList by IPv4 and IPv6 address.
func (a AddressPairList) Sort() {
	sort.Slice(a, func(i, j int) bool {
		if a[i].IPV4 < a[j].IPV4 {
			return true
		} else if a[i].IPV4 == a[j].IPV4 {
			return a[i].IPV6 < a[j].IPV6
		}
		return false
	})
}

// EndpointNetworking is the addressing information of an endpoint.
type EndpointNetworking struct {
	// IP4/6 addresses assigned to this Endpoint
	Addressing AddressPairList `json:"addressing"`

	// NodeIP is the IP of the node the endpoint is running on. The IP must
	// be reachable between nodes.
	NodeIP string `json:"node,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={dolphin},singular="dolphinnode",path="dolphinnodes",scope="Cluster",shortName={cn,dolphinn}
// +kubebuilder:printcolumn:JSONPath=".spec.addresses[?(@.type==\"DolphinInternalIP\")].ip",description="Dolphin internal IP for this node",name="DolphinInternalIP",type=string
// +kubebuilder:printcolumn:JSONPath=".spec.addresses[?(@.type==\"InternalIP\")].ip",description="IP of the node",name="InternalIP",type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",description="Time duration since creation of Dolphinnode",name="Age",type=date
// +kubebuilder:storageversion
// +kubebuilder:subresource:status

// DolphinNode represents a node managed by Dolphin. It contains a specification
// to control various node specific configuration aspects and a status section
// to represent the status of the node.
type DolphinNode struct {
	// +deepequal-gen=false
	metav1.TypeMeta `json:",inline"`
	// +deepequal-gen=false
	metav1.ObjectMeta `json:"metadata"`

	// Spec defines the desired specification/configuration of the node.
	Spec NodeSpec `json:"spec"`

	// Status defines the realized specification/configuration and status
	// of the node.
	//
	// +kubebuilder:validation:Optional
	Status NodeStatus `json:"status,omitempty"`
}

// NodeAddress is a node address.
type NodeAddress struct {
	// Type is the type of the node address
	Type addressing.AddressType `json:"type,omitempty"`

	// IP is an IP of a node
	IP string `json:"ip,omitempty"`
}

// NodeSpec is the configuration specific to a node.
type NodeSpec struct {
	// InstanceID is the identifier of the node. This is different from the
	InstanceID string `json:"instance-id,omitempty"`

	// BootID is a unique node identifier generated on boot
	//
	// +kubebuilder:validation:Optional
	BootID string `json:"bootid,omitempty"`

	// Addresses is the list of all node addresses.
	//
	// +kubebuilder:validation:Optional
	Addresses []NodeAddress `json:"addresses,omitempty"`

	// HealthAddressing is the addressing information for health connectivity
	// checking.
	//
	// +kubebuilder:validation:Optional
	HealthAddressing HealthAddressingSpec `json:"health,omitempty"`

	// IngressAddressing is the addressing information for Ingress listener.
	//
	// +kubebuilder:validation:Optional
	IngressAddressing AddressPair `json:"ingress,omitempty"`

	// Encryption is the encryption configuration of the node.
	//
	// +kubebuilder:validation:Optional
	Encryption EncryptionSpec `json:"encryption,omitempty"`

	// ENI is the AWS ENI specific configuration.
	//
	// +kubebuilder:validation:Optional
	ENI eniTypes.ENISpec `json:"eni,omitempty"`

	// Azure is the Azure IPAM specific configuration.
	//
	// +kubebuilder:validation:Optional
	Azure azureTypes.AzureSpec `json:"azure,omitempty"`

	// IPAM is the address management specification. This section can be
	//
	// +kubebuilder:validation:Optional
	IPAM ipamTypes.IPAMSpec `json:"ipam,omitempty"`

	// NodeIdentity is the Dolphin numeric identity allocated for the node, if any.
	//
	// +kubebuilder:validation:Optional
	NodeIdentity uint64 `json:"nodeidentity,omitempty"`
}

// HealthAddressingSpec is the addressing information required to do
// connectivity health checking.
type HealthAddressingSpec struct {
	// IPv4 is the IPv4 address of the IPv4 health endpoint.
	//
	// +kubebuilder:validation:Optional
	IPv4 string `json:"ipv4,omitempty"`

	// IPv6 is the IPv6 address of the IPv4 health endpoint.
	//
	// +kubebuilder:validation:Optional
	IPv6 string `json:"ipv6,omitempty"`
}

// EncryptionSpec defines the encryption relevant configuration of a node.
type EncryptionSpec struct {
	//
	// +kubebuilder:validation:Optional
	Key int `json:"key,omitempty"`
}

// NodeStatus is the status of a node.
type NodeStatus struct {
	// ENI is the AWS ENI specific status of the node.
	//
	// +kubebuilder:validation:Optional
	ENI eniTypes.ENIStatus `json:"eni,omitempty"`

	// Azure is the Azure specific status of the node.
	//
	// +kubebuilder:validation:Optional
	Azure azureTypes.AzureStatus `json:"azure,omitempty"`

	// IPAM is the IPAM status of the node.
	//
	// +kubebuilder:validation:Optional
	IPAM ipamTypes.IPAMStatus `json:"ipam,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +deepequal-gen=false

// DolphinNodeList is a list of DolphinNode objects.
type DolphinNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of DolphinNode
	Items []DolphinNode `json:"items"`
}

// InstanceID returns the InstanceID of a DolphinNode.
func (n *DolphinNode) InstanceID() (instanceID string) {
	if n != nil {
		instanceID = n.Spec.InstanceID
		// OBSOLETE: This fallback can be removed in Dolphin 1.9
		if instanceID == "" {
			instanceID = n.Spec.ENI.InstanceID
		}
	}
	return
}

func (n NodeAddress) ToString() string {
	return n.IP
}

func (n NodeAddress) AddrType() addressing.AddressType {
	return n.Type
}

// GetIP returns one of the DolphinNode's IP addresses available
func (n *DolphinNode) GetIP(ipv6 bool) net.IP {
	return addressing.ExtractNodeIP[NodeAddress](n.Spec.Addresses, ipv6)
}
