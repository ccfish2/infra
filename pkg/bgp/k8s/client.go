package k8s

import (
	"context"

	"github.com/ccfish2/infra/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client struct {
	client.Clientset
	log *logrus.Logger
}

func New(cs client.Clientset, log *logrus.Logger) *Client {
	return &Client{
		Clientset: cs,
		log:       log,
	}
}

func (c *Client) Update(svc *corev1.Service) (*corev1.Service, error) {
	return c.CoreV1().Services(svc.GetNamespace()).Update(context.TODO(), svc, metav1.UpdateOptions{})
}

func (c *Client) UpdateStatus(svc *corev1.Service) error {
	_, err := c.CoreV1().Services(svc.Namespace).UpdateStatus(context.TODO(), svc, metav1.UpdateOptions{})
	return err
}

func (c *Client) Infof(_ *corev1.Service, fmt, desc string, args ...interface{}) {
	c.log.WithField("event", desc).Infof("k8s service %s: %s", fmt, args)
}

func (c *Client) Errorf(_ *corev1.Service, fmt, desc string, args ...interface{}) {
	c.log.WithField("event", desc).Errorf(fmt, args...)
}
