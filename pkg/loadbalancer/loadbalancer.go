package loadbalancer

import (
	"strings"
	"unsafe"

	"github.com/ccfish2/infra/pkg/container/cache"
	"github.com/ccfish2/infra/statedb/index"
	"github.com/cespare/xxhash/v2"
)

type L4Type = string

// +deepequal-gen=true
// +deepequal-gen:private-method=true
type L4Addr struct {
	Protocol L4Type
	Port     uint16
}

// ServiceName represents the fully-qualified reference to the service by name,
// including both the namespace and name of the service (and optionally the cluster).
// +k8s:deepcopy-gen=true
type ServiceName struct {
	// str is (<cluster>/)<namespace>/<name>
	str string

	// namePos is where the name starts
	// (<cluster>/)<namespace>/<name>
	//                         ^
	namePos uint16

	// clusterEndPos is where the cluster (including '/' ends. If zero then there is
	// no cluster.
	// (<cluster>/)<namespace>/<name>
	//             ^
	clusterEndPos uint16
}

func (s *ServiceName) DeepEqual(other *ServiceName) bool {
	switch {
	case s == nil && other == nil:
		return true
	case s != nil && other != nil:
		return *s == *other
	default:
		return false
	}
}

func (s ServiceName) Cluster() string {
	if s.clusterEndPos > 0 {
		return s.str[:s.clusterEndPos-1]
	}
	return ""
}

func (s ServiceName) Name() string {
	return s.str[s.namePos:]
}

func (s ServiceName) Namespace() string {
	return s.str[s.clusterEndPos : s.namePos-1]
}

func (n ServiceName) Key() index.Key {
	// index.Key is never mutated so it's safe to return the underlying
	// string as []byte without copying.
	return unsafe.Slice(unsafe.StringData(n.str), len(n.str))
}

// serviceNameCache for deduplicating the [ServiceName.str] to reduce overall
// memory usage.
var serviceNameCache = cache.New(
	func(n ServiceName) uint64 {
		return serviceNameHash(n.Cluster(), n.Namespace(), n.Name())
	},
	nil,
	func(a, b ServiceName) bool {
		return b.str != "" /* only match if non-zero value */ &&
			a.str == b.str
	},
)

func serviceNameHash(cluster, namespace, name string) uint64 {
	var d xxhash.Digest
	d.WriteString(cluster)
	d.WriteString(namespace)
	d.WriteString(name)
	return d.Sum64()
}

func NewServiceName(namespace, name string) ServiceName {
	return NewServiceNameInCluster("", namespace, name)
}

func NewServiceNameInCluster(cluster, namespace, name string) ServiceName {
	return cache.GetOrPutWith(
		serviceNameCache,
		serviceNameHash(cluster, namespace, name),
		func(sn ServiceName) bool {
			return len(sn.str) > 0 &&
				sn.Cluster() == cluster && sn.Namespace() == namespace && sn.Name() == name
		},
		func() ServiceName {
			// ServiceName not found from cache, create it.
			var b strings.Builder
			pos := 0
			if cluster != "" {
				n, _ := b.WriteString(cluster)
				b.WriteRune('/')
				pos += n + 1
			}
			cendPos := pos
			n, _ := b.WriteString(namespace)
			b.WriteRune('/')
			pos += n + 1
			b.WriteString(name)
			return ServiceName{
				str:           b.String(),
				clusterEndPos: uint16(cendPos),
				namePos:       uint16(pos),
			}
		},
	)
}
