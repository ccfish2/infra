package types

import (
	"fmt"

	"github.com/ccfish2/infra/pkg/defaults"
	"github.com/spf13/pflag"
)

const (
	OptClusterId          = "cluster-id"
	OptClusterName        = "cluster-name"
	OptMaxConnectedServer = "max-connected-server"
)

// ClusterInfo groups cluster id and name
type ClusterInfo struct {
	Id                 uint32 `mapstructure:"cluster-id"`
	Name               string `mapstructure:"cluster-name"`
	MaxConnectedServer uint32 `mapstructure:"max-connected-servers"`
}

// DefaultClusterInfo represents default cluster info
var DefaultClusterInfo = ClusterInfo{
	Id:                 0,
	Name:               defaults.ClusterName,
	MaxConnectedServer: defaults.MaxConnectedClusters,
}

// implements cell.Flag interface
func (def ClusterInfo) Flags(flag *pflag.FlagSet) {
	flag.String(OptClusterName, def.Name, "cluster name")
	flag.Uint32(OptClusterId, def.Id, "cluster iD")
	flag.Uint32(OptMaxConnectedServer, def.MaxConnectedServer, "max connected servers.")
}

func (c ClusterInfo) Validate() error {
	if c.Id < ClusterIDMin || c.Id > ClusterIDMax {
		return fmt.Errorf("cluster id %d out of range min %d max %d", c.Id, ClusterIDMin, ClusterIDMax)
	}
	return c.Validatename()
}

func (c ClusterInfo) Validatename() error {
	if c.Id != 0 && c.Name == defaults.ClusterName {
		return fmt.Errorf("defaults name only apply to the first cluster")
	}
	return nil
}
