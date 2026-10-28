package v1

import (
	"encoding/json"
	"net"
	"testing"

	eniTypes "github.com/ccfish2/infra/pkg/aws/eni/types"
	"github.com/ccfish2/infra/pkg/node/addressing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestAddressPairListSort(t *testing.T) {
	addressPairs := AddressPairList{
		{IPV4: "10.0.0.20", IPV6: "2001:db8::20"},
		{IPV4: "10.0.0.10", IPV6: "2001:db8::20"},
		{IPV4: "10.0.0.10", IPV6: "2001:db8::10"},
	}

	addressPairs.Sort()

	assert.Equal(t, AddressPairList{
		{IPV4: "10.0.0.10", IPV6: "2001:db8::10"},
		{IPV4: "10.0.0.10", IPV6: "2001:db8::20"},
		{IPV4: "10.0.0.20", IPV6: "2001:db8::20"},
	}, addressPairs)
}

func TestServiceListenerServiceName(t *testing.T) {
	listener := &ServiceListener{Name: "frontend", Namespace: "production"}
	serviceName := listener.ServiceName()

	assert.Equal(t, "frontend", serviceName.Name())
	assert.Equal(t, "production", serviceName.Namespace())
	assert.Empty(t, serviceName.Cluster())
}

func TestDolphinNodeInstanceID(t *testing.T) {
	tests := []struct {
		name     string
		node     *DolphinNode
		expected string
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: "",
		},
		{
			name: "primary instance ID",
			node: &DolphinNode{Spec: NodeSpec{
				InstanceID: "i-primary",
				ENI:        eniTypes.ENISpec{InstanceID: "i-legacy"},
			}},
			expected: "i-primary",
		},
		{
			name: "legacy ENI instance ID fallback",
			node: &DolphinNode{Spec: NodeSpec{
				ENI: eniTypes.ENISpec{InstanceID: "i-legacy"},
			}},
			expected: "i-legacy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.node.InstanceID())
		})
	}
}

func TestNodeAddress(t *testing.T) {
	address := NodeAddress{Type: addressing.NodeInternalIP, IP: "10.0.0.10/32"}

	assert.Equal(t, "10.0.0.10/32", address.ToString())
	assert.Equal(t, addressing.NodeInternalIP, address.AddrType())
}

func TestDolphinNodeGetIP(t *testing.T) {
	node := &DolphinNode{Spec: NodeSpec{Addresses: []NodeAddress{
		{Type: addressing.NodeExternalIP, IP: "192.0.2.10/32"},
		{Type: addressing.NodeInternalIP, IP: "10.0.0.10/32"},
		{Type: addressing.NodeExternalIP, IP: "2001:db8::10/128"},
		{Type: addressing.NodeInternalIP, IP: "2001:db8::20/128"},
	}}}

	ipv4 := node.GetIP(false)
	ipv6 := node.GetIP(true)

	assert.True(t, ipv4.Equal(net.ParseIP("10.0.0.10")))
	assert.True(t, ipv6.Equal(net.ParseIP("2001:db8::20")))
	assert.Nil(t, (&DolphinNode{Spec: NodeSpec{Addresses: []NodeAddress{
		{Type: addressing.NodeInternalIP, IP: "not-a-cidr"},
	}}}).GetIP(false))
}

func TestXDSResourceJSONRoundTrip(t *testing.T) {
	value := wrapperspb.String("test-resource")
	resourceAny, err := anypb.New(value)
	require.NoError(t, err)

	original := &XDSResource{Any: resourceAny}
	encoded, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded XDSResource
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	assert.True(t, original.DeepEqual(&decoded))
}

func TestXDSResourceDeepCopyAndDeepEqual(t *testing.T) {
	resourceAny, err := anypb.New(wrapperspb.String("copy-me"))
	require.NoError(t, err)
	original := &XDSResource{Any: resourceAny}

	copy := &XDSResource{}
	original.DeepCopyInto(copy)

	assert.True(t, original.DeepEqual(copy))
	assert.NotSame(t, original.Any, copy.Any)
}

func TestXDSResourceUnmarshalInvalidJSON(t *testing.T) {
	var resource XDSResource

	// Invalid JSON is intentionally tolerated by this API type and reported through
	// the containing resource decoder instead of returning a partially decoded Any.
	assert.NoError(t, json.Unmarshal([]byte(`{"not":"an xDS resource"}`), &resource))
}
