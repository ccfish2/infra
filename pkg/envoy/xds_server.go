package envoy

import (
	envoy_config_core "github.com/cilium/proxy/go/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func GetInternalListenerCIDRs(ipv4, ipv6 bool) []*envoy_config_core.CidrRange {
	var cidrRanges []*envoy_config_core.CidrRange
	if ipv4 {
		cidrRanges = append(cidrRanges,
			[]*envoy_config_core.CidrRange{
				{AddressPrefix: "10.0.0.0", PrefixLen: &wrapperspb.UInt32Value{Value: 8}},
				{AddressPrefix: "172.16.0.0", PrefixLen: &wrapperspb.UInt32Value{Value: 12}},
				{AddressPrefix: "192.168.0.0", PrefixLen: &wrapperspb.UInt32Value{Value: 16}},
				{AddressPrefix: "127.0.0.1", PrefixLen: &wrapperspb.UInt32Value{Value: 32}}}...)
	}
	if ipv6 {
		cidrRanges = append(cidrRanges, &envoy_config_core.CidrRange{
			AddressPrefix: "::1",
			PrefixLen:     &wrapperspb.UInt32Value{Value: 128},
		})
	}
	return cidrRanges
}
