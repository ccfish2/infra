package option

const (
	IPAMKubernetes = "kubernetes"
	IPAMCRD        = "crd"
	IPAMENI        = "eni"
	IPAMAzure      = "azure"

	IPAMClusterPool     = "cluster-pool"
	IPAMMultiPool       = "multi-pool"
	IPAMAlibabaCloud    = "alibabacloud"
	IPAMDelegatedPlugin = "delegated-plugin"
)

const (
	IPAMMarkForRelease  = "marked-for-release"
	IPAMReadyForRelease = "ready-for-release"
	IPAMDoNotRelease    = "do-not-release"
	IPAMReleased        = "released"
)

const ENIPDBlockSizeIPv4 = 16
