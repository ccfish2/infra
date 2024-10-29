package defaults

import "time"

const (
	IPv6ClusterAllocCIDRBase = "f00d::"
	IPv6ClusterAllocCIDR     = IPv6ClusterAllocCIDRBase + "/64"

	IPAMDefaultIPPool               = "default"
	EnableHostIPRestore             = true
	EnableHealthChecking            = true
	EnableEndpointHealthChecking    = true
	EnableHealthCheckNodePort       = true
	EnableHealthCheckLoadBalancerIP = false
	EnableIPv4                      = true
	EnableIPv6                      = true
	EnableIPv6NDP                   = false
	EnableSRv6                      = false
	EnableL7Proxy                   = true

	EnableSCTP = false

	DNSMaxIPsPerRestoredRule = 1000
	ToFQDNsMaxIPsPerHost     = 50

	KVstorePeriodicSync        = 5 * time.Minute
	KVstoreConnectivityTimeout = 2 * time.Minute
	IPAllocationTimeout        = 2 * time.Minute

	IdentityChangeGracePeriod  = 5 * time.Second
	IdentityRestoreGracePeriod = 10 * time.Minute

	LoopbackIPv4              = "169.254.42.1"
	EnableEndpointRoutes      = false
	AnnotateK8sNode           = false
	K8sServiceCacheSize       = 128
	AllowICMPFragNeeded       = true
	EnableWellKnownIdentities = true
	AllocatorListTimeout      = 3 * time.Minute
	EnableICMPRules           = true
	ExternalClusterIP         = false
	EnableVTEP                = false
	EnableBGPControlPlane     = false
	EnableK8sNetworkPolicy    = true
	EnableEnvoyConfig         = false
)
