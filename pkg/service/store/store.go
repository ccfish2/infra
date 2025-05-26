package store

import (
	"github.com/ccfish2/infra/pkg/loadbalancer"
)

// +deepequal-gen=true
type PortConfiguration map[string]*loadbalancer.L4Addr
