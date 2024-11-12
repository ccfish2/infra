package resource

import (
	dolphin_v1 "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var scheme = runtime.NewScheme()

var localSchemeBuilder = runtime.SchemeBuilder{
	dolphin_v1.AddToScheme,
}

var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	utilruntime.Must(AddToScheme(scheme))
}
