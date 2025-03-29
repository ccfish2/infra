package mock

import (
	v1 "k8s.io/api/core/v1"

	"github.com/ccfish2/infra/pkg/k8s"
	metallbspr "github.com/ccfish2/metalb0110/pkg/speaker"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenTestServiceParis() (v1.Service, v1.Service, metallbspr.Service, k8s.ServiceID) {
	const (
		IP = "10.10.10.10"
	)
	spec := v1.ServiceSpec{
		Type:                  "TestType",
		ExternalTrafficPolicy: "TestExternalTrafficPolicy",
	}
	meta := metav1.ObjectMeta{
		Name:            "TestName",
		Namespace:       "TestNamespace",
		ResourceVersion: "TestResourceVersion",
	}
	ing := v1.LoadBalancerIngress{
		IP:       IP,
		Hostname: "",
		Ports:    nil,
	}
	var ing v1.LoadBalancerIngress
	lbStatus := v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{ing},
	}
	status := v1.ServiceStatus{
		LoadBalancer: lbStatus,
	}
	svc := v1.Service{
		Spec:       spec,
		Status:     status,
		ObjectMeta: meta,
	}
	metallbSvc := metallbspr.Service{
		Type:          string(spec.Type),
		TrafficPolicy: string(spec.ExternalTrafficPolicy),
		Ingress:       []v1.LoadBalancerIngress{},
	}
	V1sERVICE := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "TestName",
			Namespace:       "TestNamespace",
			ResourceVersion: "TestResourceVersion",
		},
		Spec: v1.ServiceSpec{
			Type:                  "TestType",
			ExternalTrafficPolicy: "TestExternalTrafficPolicy",
		},
		Status: v1.ServiceStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{IP: IP,
						Ports: nil},
				},
			},
		},
	}

	serviceiD := k8s.ServiceID{
		Name:      V1sERVICE.Name,
		Namespace: V1sERVICE.Namespace,
	}
	return svc, V1sERVICE, metallbSvc, serviceiD
}
