package v2alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GroupVersion       = schema.GroupVersion{Group: "dolphin.io", Version: "v2alpha1"}
	SchemeGroupVersion = GroupVersion
)
