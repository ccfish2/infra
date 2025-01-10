package k8s

import (
	"github.com/ccfish2/infra/pkg/hive/cell"
	dolphin_api_v1 "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io/v1"
	"github.com/ccfish2/infra/pkg/k8s/client"
	"github.com/ccfish2/infra/pkg/k8s/resource"
	"github.com/ccfish2/infra/pkg/k8s/utils"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func DolphinIdentityResource(lc cell.Lifecycle, cs client.Clientset, opts ...func(*metav1.ListOptions)) (resource.Resource[*dolphin_api_v1.DolphinIdentity], error) {
	if !cs.IsEnabled() {
		return nil, nil
	}
	lw := utils.ListerWatcherWithModifiers(
		utils.ListerWatcherFromTyped[*dolphin_api_v1.DolphinIdentityList](cs.DolphinV1().DolphinIdentities()),
		opts...,
	)
	return resource.New[*dolphin_api_v1.DolphinIdentity](lc, lw, resource.WithMetric("DolphinIdentityList")), nil
}

func DolphinNodeResource(lc cell.Lifecycle, cs client.Clientset, opts ...func(*metav1.ListOptions)) (resource.Resource[*dolphin_api_v1.DolphinNode], error) {
	if !cs.IsEnabled() {
		return nil, nil
	}
	lw := utils.ListerWatcherWithModifiers(
		utils.ListerWatcherFromTyped[*dolphin_api_v1.DolphinNodeList](cs.DolphinV1().DolphinNodes()),
		opts...,
	)
	return resource.New[*dolphin_api_v1.DolphinNode](lc, lw, resource.WithMetric("DolphinNode")), nil
}
