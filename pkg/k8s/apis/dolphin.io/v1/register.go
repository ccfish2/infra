package v1

import (
	apisgroupconst "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	CustomResourceDefinitionGroup   = apisgroupconst.CustomResourceDefinitionGroup
	CustomResourceDefinitionVersion = "v1"

	DEPPluralName     = "dolphinendpoints"
	DEPKindDefinition = "DolphinEndpoint"
	DEPName           = DEPPluralName + "." + CustomResourceDefinitionGroup

	DEPSPluralName     = "dolphinendpointslices"
	DEPSKindDefinition = "DolphinEndpointSlice"
	DEPSName           = DEPSPluralName + "." + CustomResourceDefinitionGroup

	DECPluralName     = "dolphinenvoyconfigs"
	DECKindDefinition = "DolphinEnvoyConfig"
	DECName           = DECPluralName + "." + CustomResourceDefinitionGroup

	DIDPluralName     = "dolphinidentities"
	DIDKindDefinition = "DolphinIdentity"
	DIDName           = DIDPluralName + "." + CustomResourceDefinitionGroup
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   CustomResourceDefinitionGroup,
	Version: CustomResourceDefinitionVersion,
}

var (
	// SchemeBuilder is needed by DeepCopy generator.
	SchemeBuilder runtime.SchemeBuilder
	// localSchemeBuilder and AddToScheme will stay in k8s.io/kubernetes.
	localSchemeBuilder = &SchemeBuilder

	// AddToScheme adds all types of this clientset into the given scheme.
	// This allows composition of clientsets, like in:
	//
	//   import (
	//     "k8s.io/client-go/kubernetes"
	//     clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	//     aggregatorclientsetscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
	//   )
	//
	//   kclientset, _ := kubernetes.NewForConfig(c)
	//   aggregatorclientsetscheme.AddToScheme(clientsetscheme.Scheme)
	AddToScheme = localSchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&DolphinEndpoint{},
		&DolphinEndpointList{},
		&DolphinEnvoyConfig{},
		&DolphinEnvoyConfigList{},
		&DolphinIdentity{},
		&DolphinIdentityList{},
		&DolphinEndpointSlice{},
		&DolphinEndpointSliceList{})

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
