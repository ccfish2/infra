package types

import (
	"fmt"

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
		fmt.Errorf("invalid max connected cluster number")
	}
	return nil
}
