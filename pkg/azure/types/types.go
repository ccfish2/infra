package types

type AzureSpec struct {
	// +kubebuilder:validation:Optional
	InterfaceName string `json:"interface-name,omitempty"`
}

// AzureStatus is the status of the Azure IPAM
// +k8s:deepcpy-gen=true
type AzureStatus struct {
	Interfaces []AzureInterface `json:"interfaces,omitempty"`
}

type AzureAddress struct {
	IP     string `json:"ip,omitempty"`
	Subnet string `json:"subnet,omitempty"`
	State  string `json:"state,omitempty"`
}

// AzureInterface is the interface of the Azure Interface
//
// +k8s:deepcpy-gen=true
type AzureInterface struct {
	ID            string         `json:"id,omitempty"`
	Name          string         `json:"name,omitempty"`
	MAC           string         `json:"mac,omitempty"`
	State         string         `json:"state,omitempty"`
	Addresses     []AzureAddress `json:"addresses,omitempty"`
	SecurityGroup string         `json:"security-group,omitempty"`
	GatewayIP     string         `json:"GatewayIP"`
	Gateway       string         `json:"gateway"`
	CIDR          string         `json:"cidr,omitempty"`
	vmssName      string         `json:"-"`
	vmID          string         `json:"-"`
	resourceGroup string         `json:"-"`
}
