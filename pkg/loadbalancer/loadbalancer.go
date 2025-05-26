package loadbalancer

type L4Type = string

// +deepequal-gen=true
// +deepequal-gen:private-method=true
type L4Addr struct {
	Protocol L4Type
	Port     uint16
}
