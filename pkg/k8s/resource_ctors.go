package k8s

import "github.com/spf13/pflag"

type Config struct {
	EnableK8sEndpointSlice bool
	K8sServiceProxyName    string
}

var DefaultConfig = Config{
	EnableK8sEndpointSlice: true,
}

func (def Config) Flags(flags *pflag.FlagSet) {
	flags.Bool("enable-k8s-endpoint-slice", def.EnableK8sEndpointSlice, "Enables k8s EndpointSlice feature in Dolphin if the k8s cluster supports it")
	flags.String("k8s-service-proxy-name", def.K8sServiceProxyName, "Value of K8s service-proxy-name label for which Dolphin handles the services (empty = all services without service.kubernetes.io/service-proxy-name label)")
}
