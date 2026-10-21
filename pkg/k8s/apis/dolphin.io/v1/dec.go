package v1

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ccfish2/infra/pkg/loadbalancer"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/ccfish2/infra/pkg/option"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={dolphin},singular="dolphinenvoyconfig",path="dolphinenvoyconfigs",scope="Namespaced",shortName={dec}
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",description="The age of the DolphinEnvoyConfig",name="Age",type=date
// +kubebuilder:storageversion

type DolphinEnvoyConfig struct {
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	// +kubebuilder:validation:Required
	metav1.ObjectMeta `json:"metadata"`

	// +k8s:openapi-gen=false
	// +kubebuilder:validation:Optional
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
	// +kubebuilder:validation:Optional
	Services []*ServiceListener `json:"services,omitempty"`

	// +kubebuilder:validation:Optional
	BackendServices []*Service `json:"backendServices,omitempty"`

	// +kubebuilder:validation:Optional
	Resources []XDSResource `json:"resources,omitempty"`

	// +kubebuilder:validation:Optional
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`
}

type Service struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:validation:Optional
	Ports []string `json:"number,omitempty"`
}

type ServiceListener struct {

	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:validation:Optional
	Ports []uint16 `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	Listener string `json:"listener,omitempty"`
}

func (l *ServiceListener) ServiceName() loadbalancer.ServiceName {
	return loadbalancer.NewServiceName(l.Namespace, l.Name)
}

// +kubebuilder:pruning:PreserveUnknownFields
type XDSResource struct {
	*anypb.Any `json:"-"`
}

// DeepCopyInto deep copies 'in' into 'out'.
func (in *XDSResource) DeepCopyInto(out *XDSResource) {
	out.Any, _ = proto.Clone(in.Any).(*anypb.Any)
}

// DeepEqual returns 'true' if 'a' and 'b' are equal.
func (a *XDSResource) DeepEqual(b *XDSResource) bool {
	return proto.Equal(a.Any, b.Any)
}

// MarshalJSON ensures that the unstructured object produces proper
// JSON when passed to Go's standard JSON library.
func (u *XDSResource) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(u.Any)
}

// UnmarshalJSON ensures that the unstructured object properly decodes
// JSON when passed to Go's standard JSON library.
func (u *XDSResource) UnmarshalJSON(b []byte) (err error) {
	// xDS resources are not validated in K8s, recover from possible panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("DEC JSON decoding paniced: %v", r)
		}
	}()
	u.Any = &anypb.Any{}
	err = protojson.Unmarshal(b, u.Any)
	if err != nil {
		var buf bytes.Buffer
		json.Indent(&buf, b, "", "\t")
		// slogloggercheck: it's safe to use the default logger here as it has been initialized by the program up to this point.
		logging.DefaultSlogLogger.Warn("Ignoring invalid DolphinEnvoyConfig JSON",
			logfields.Error, err,
			logfields.Object, buf,
		)
	} else if option.Config.Debug {
		// slogloggercheck: it's safe to use the default logger here as it has been initialized by the program up to this point.
		logging.DefaultSlogLogger.Debug("DEC unmarshaled XDS Resource", logfields.Resource, prototext.Format(u.Any))
	}
	return nil
}
