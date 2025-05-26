package k8s

import "fmt"

type ServiceID struct {
	Cluster   string `json:"cluster,omitempty"`
	Name      string `json:"serviceName,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func (s ServiceID) String() string {
	if s.Cluster != "" {
		return fmt.Sprintf("%s/%s/%s", s.Cluster, s.Namespace, s.Name)
	}
	return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
}

type EndpointSliceID struct {
	ServiceID
	EndpointSliceName string
}
