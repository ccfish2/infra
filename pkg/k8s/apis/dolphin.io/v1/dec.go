package v1

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={dolphin},singular="dolphinenvoyconfig",path="dolphinenvoyconfigs",scope="Namespaced",shortName={dec,dolphinec}
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

func (in *XDSResource) DeepCopyInto(out *XDSResource) {
	out.Any, _ = proto.Clone(in.Any).(*anypb.Any)
}

func (in *XDSResource) DeepEqual(other *XDSResource) bool {
	return proto.Equal(in.Any, other.Any)
}

func (in *XDSResource) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(in.Any)
}

func (in *XDSResource) UnmarshalJSON(b []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("json decoding panic")
		}
	}()
	var o anypb.Any
	if err := protojson.Unmarshal(b, &o); err != nil {
		var buf bytes.Buffer
		json.Indent(&buf, b, "", "\t")
		err = fmt.Errorf(" error %s", buf.String())
	}
	return nil
}
