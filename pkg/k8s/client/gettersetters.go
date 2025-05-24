package client

import (
	"context"

	dolphinv1 "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io/v1"
	corev1 "k8s.io/api/core/v1"
)

type Getters interface {
	GetSecrets(ctx context.Context, namespace, name string) (map[string][]byte, error)
	GetK8sNode(ctx context.Context, nodeName string) (*corev1.Node, error)
	GetDolphinNode(ctx context.Context, nodeName string) (*dolphinv1.DolphinNode, error)
}

// clientsetGetters implements the Getters interface in terms of the clientset.
type clientsetGetters struct {
	Clientset
}
