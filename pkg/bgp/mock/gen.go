package mock

import (
	v1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/ccfish2/infra/pkg/k8s"
	metallbspr "go.universe.tf/metallb/pkg/speaker"
)

func GenTestServiceParis() (slim_corev1.Service, v1.Service, metallbspr.Service, k8s.ServiceID) {
	panic("generating those services.")
}
