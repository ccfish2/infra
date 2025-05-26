package types

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +k8s:deepcopy-gen=false
type UnserializableObject struct{}

func (UnserializableObject) GetObjectKind() schema.ObjectKind {
	// Not serializable, so return the empty kind.
	return schema.EmptyObjectKind
}
