package annotation

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Get(obj metav1.Object, key string, alias ...string) (value string, ok bool) {
	keys := append([]string{key}, alias...)
	for _, k := range keys {
		if val, ok := obj.GetAnnotations()[k]; ok {
			return val, ok
		}
	}
	return "", false
}
