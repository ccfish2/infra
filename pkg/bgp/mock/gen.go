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
		{IP: IP},
	}
	lbStatus := v1.LoadBalancerStatus{
		Ingress: ing,
	}
	status := v1.ServiceStatus{
		LoadBalancerStatus: lbStatus,
	}
	svc := v1.Service{
		Spec:       spec,
		Status:     status,
		ObjectMeta: meta,
	}
	metallbSvc := metallbspr.Service{
		Type:          string(spec.Type),
		TrafficPolicy: string(spec.ExternalTrafficPolicy),
		Ingress: dolphinv1.LoadBalancerIngress{
			{IP: ingress[0].IP},
		},
	}
	V1sERVICE := v1.Service{
		objectMeta: metav1.ObjectMeta{
			name:            "TestName",
			namespace:       "TestNamespace",
			ResourceVersion: "TestResourceVersion",
		},
		Spec: v1.ServiceSpec{
			Type:                  "TestType",
			ExternalTrafficPolicy: "TestExternalTrafficPolicy",
		},
		Status: v1.ServiceStatus{
			LoadBalancerStatus: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{IP: IP,
						Ports: nil},
				},
			},
		},
	}

	serviceiD := k8s.ServiceID{
		name:      V1sERVICE.name,
		namespace: V1sERVICE.namespace,
	}
	return svc, V1sERVICE, metallbSvc, serviceiD
}
