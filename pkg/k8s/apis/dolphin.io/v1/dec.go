package v1

import (
	"google.golang.org/protobuf/types/known/anypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={dolphin},singular="dolphinenvoyconfig",path="dolphinenvoyconfigs",scope="Namespaced",shortName={dec}
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",description="The age of the identity",name="Age",type=date
// +kubebuilder:storageversion

type DolphinEnvoyConfig struct {
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	metav1.ObjectMeta `json:"metadata"`

	// +k8s:openapi-gen=false
	Spec DolphinEnvoyConfigSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +deepequal-gen=false

// DolphinEnvoyConfigList is a list of DolphinEnvoyConfig objects.
type DolphinEnvoyConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of DolphinEnvoyConfig.
	Items []DolphinEnvoyConfig `json:"items"`
}

type DolphinEnvoyConfigSpec struct {
	Services        []*ServiceListener `json:"services,omitempty"`
	BackendServices []*Service         `json:"backendServices,omitempty"`
	Resources       []XDSResource      `json:"resources,omitempty"`
}
type Service struct {
	Name      string   `json:"name,omitempty"`
	Namespace string   `json:"namespace,omitempty"`
	Ports     []string `json:"number,omitempty"`
}
type ServiceListener struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Listener  string `json:"listener,omitempty"`
}

type XDSResource struct {
	*anypb.Any `json:"-"`
}
