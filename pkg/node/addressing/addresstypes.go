package addressing

import "net"

type AddressType string

const (
	NodeHostName          AddressType = "HostName"
	NodeExternalIP        AddressType = "ExternalIP"
	NodeInternalIP        AddressType = "InternalIP"
	NodeExternalDNS       AddressType = "ExternalDNS"
	NodeInternalDNS       AddressType = "InternalDNS"
	NodeDolphinInternalIP AddressType = "DolphinInternalIP"
)

type Address interface {
	AddrType() AddressType
	ToString() string
}

func ExtractNodeIP[T Address](addrs []T, ipv6 bool) net.IP {
	var backupIP net.IP
	for _, addr := range addrs {
		parsed, _, err := net.ParseCIDR(addr.ToString())
		if parsed == nil || err != nil {
			continue
		}
		if (ipv6 && parsed.To4() == nil) || (!ipv6 && parsed.To4() != nil) {
			continue
		}
		switch addr.AddrType() {
		case NodeInternalIP:
			return parsed
		case NodeExternalIP:
			// fallback to ExternalIP if InternalIP is not found
			backupIP = parsed
		case NodeDolphinInternalIP:
			continue
		default:
			if backupIP == nil {
				backupIP = parsed
			}
		}
	}
	return backupIP
}
