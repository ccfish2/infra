package k8s

// ServiceID identifies the Kubernetes service
type ServiceID struct {
	Cluster   string `json:"cluster,omitempty"`
	Name      string `json:"serviceName,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}
