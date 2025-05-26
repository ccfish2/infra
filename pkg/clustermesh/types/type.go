package types

import (
	"fmt"

	"net/netip"

	"github.com/ccfish2/infra/pkg/defaults"
)

const (
	ClusterIDMin  = 0
	ClusterExt511 = 511
)

var ClusterIDMax uint32 = defaults.MaxConnectedClusters

func (c ClusterInfo) InitClusterIDMax() error {
	switch c.MaxConnectedServer {
	case defaults.MaxConnectedClusters, ClusterExt511:
		ClusterIDMax = c.MaxConnectedServer
	default:
		return fmt.Errorf("--%s=%d is invalid; supported values are [%d, %d]", OptMaxConnectedServer, c.MaxConnectedServer, defaults.MaxConnectedClusters, ClusterExt511)
	}
	return nil
}

func ValidateClusterID(id uint32) error {
	if id <= ClusterIDMin || id > ClusterIDMax {
		return fmt.Errorf("cluster id %d out of range min %d max %d", id, ClusterIDMin, ClusterIDMax)
	}
	return nil
}

type AddrCluster struct {
	addr      netip.Addr
	clusterID uint32
}
