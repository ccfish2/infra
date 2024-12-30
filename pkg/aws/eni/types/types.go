package types

import "github.com/ccfish2/infra/pkg/ipam/types"

// ENISpec is the ENI specification of a node
type ENISpec struct {
	// InstanceID is the AWS InstanceId of the node
	InstanceID string `json:"instance-id,omitempty"`

	// InstanceType is the AWS EC2 instance type, e.g. "m5.large"
	InstanceType string `json:"instance-type,omitempty"`

	// MinAllocate is the minimum number of IPs that must be allocated when
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	MinAllocate int `json:"min-allocate,omitempty"`

	// PreAllocate defines the number of IP addresses that must
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	PreAllocate int `json:"pre-allocate,omitempty"`

	// MaxAboveWatermark is the maximum number of addresses to allocate
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	MaxAboveWatermark int `json:"max-above-watermark,omitempty"`

	// FirstInterfaceIndex is the index of the first ENI to use for IP
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	FirstInterfaceIndex *int `json:"first-interface-index,omitempty"`

	// SecurityGroups is the list of security groups to attach to any ENI
	// +kubebuilder:validation:Optional
	SecurityGroups []string `json:"security-groups,omitempty"`

	// SecurityGroupTags is the list of tags to use when evaliating what
	// +kubebuilder:validation:Optional
	SecurityGroupTags map[string]string `json:"security-group-tags,omitempty"`

	// SubnetIDs is the list of subnet ids to use when evaluating what AWS
	// +kubebuilder:validation:Optional
	SubnetIDs []string `json:"subnet-ids,omitempty"`

	// SubnetTags is the list of tags to use when evaluating what AWS
	// +kubebuilder:validation:Optional
	SubnetTags map[string]string `json:"subnet-tags,omitempty"`

	// NodeSubnetID is the subnet of the primary ENI the instance was brought up
	// +kubebuilder:validation:Optional
	NodeSubnetID string `json:"node-subnet-id,omitempty"`

	// VpcID is the VPC ID to use when allocating ENIs.
	//
	// +kubebuilder:validation:Optional
	VpcID string `json:"vpc-id,omitempty"`

	// AvailabilityZone is the availability zone to use when allocating
	// +kubebuilder:validation:Optional
	AvailabilityZone string `json:"availability-zone,omitempty"`

	// ExcludeInterfaceTags is the list of tags to use when excluding ENIs
	// +kubebuilder:validation:Optional
	ExcludeInterfaceTags map[string]string `json:"exclude-interface-tags,omitempty"`

	// DeleteOnTermination defines that the ENI should be deleted
	// +kubebuilder:validation:Optional
	DeleteOnTermination *bool `json:"delete-on-termination,omitempty"`

	// UsePrimaryAddress determines whether an ENI's primary address
	// +kubebuilder:validation:Optional
	UsePrimaryAddress *bool `json:"use-primary-address,omitempty"`

	// DisablePrefixDelegation determines whether ENI prefix delegation should be
	// disabled on this node.
	//
	// +kubebuilder:validation:Optional
	DisablePrefixDelegation *bool `json:"disable-prefix-delegation,omitempty"`
}

// ENI represents an AWS Elastic Network Interface
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html
type ENI struct {
	// +optional
	ID string `json:"id,omitempty"`

	// +optional
	IP string `json:"ip,omitempty"`
	// +optional
	MAC string `json:"mac,omitempty"`
	// +optional
	AvailabilityZone string `json:"availability-zone,omitempty"`
	// +optional
	Description string `json:"description,omitempty"`

	// +optional
	Number int `json:"number,omitempty"`

	// +optional
	Subnet AwsSubnet `json:"subnet,omitempty"`

	// +optional
	VPC AwsVPC `json:"vpc,omitempty"`

	// +optional
	Addresses []string `json:"addresses,omitempty"`

	// +optional
	Prefixes []string `json:"prefixes,omitempty"`

	// SecurityGroups are the security groups associated with the ENI
	SecurityGroups []string `json:"security-groups,omitempty"`

	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

func (e *ENI) DeepCopyInterface() types.Interface {
	return e.DeepCopy()
}

// InterfaceID returns the identifier of the interface
func (e *ENI) InterfaceID() string {
	return e.ID
}

// ForeachAddress iterates over all addresses and calls fn
func (e *ENI) ForeachAddress(id string, fn types.AddressIterator) error {
	for _, address := range e.Addresses {
		if err := fn(id, e.ID, address, "", address); err != nil {
			return err
		}
	}

	return nil
}

// IsExcludedBySpec returns true if the ENI is excluded by the provided spec and
func (e *ENI) IsExcludedBySpec(spec ENISpec) bool {
	if spec.FirstInterfaceIndex != nil && e.Number < *spec.FirstInterfaceIndex {
		return true
	}

	if len(spec.ExcludeInterfaceTags) > 0 {
		if types.Tags(e.Tags).Match(spec.ExcludeInterfaceTags) {
			return true
		}
	}

	return false
}

// ENIStatus is the status of ENI addressing of the node
type ENIStatus struct {
	// ENIs is the list of ENIs on the node
	//
	// +optional
	ENIs map[string]ENI `json:"enis,omitempty"`
}

// AwsSubnet stores information regarding an AWS subnet
type AwsSubnet struct {
	// ID is the ID of the subnet
	ID string `json:"id,omitempty"`

	// CIDR is the CIDR range associated with the subnet
	CIDR string `json:"cidr,omitempty"`
}

// AwsVPC stores information regarding an AWS VPC
type AwsVPC struct {
	/// ID is the ID of a VPC
	ID string `json:"id,omitempty"`

	// PrimaryCIDR is the primary CIDR of the VPC
	PrimaryCIDR string `json:"primary-cidr,omitempty"`

	// CIDRs is the list of CIDR ranges associated with the VPC
	CIDRs []string `json:"cidrs,omitempty"`
}
