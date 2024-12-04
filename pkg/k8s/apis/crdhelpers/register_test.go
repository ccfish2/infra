package crdhelpers

import (
	"testing"
	"time"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// myself
	k8sconst "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io"
	"github.com/ccfish2/infra/pkg/versioncheck"
	"github.com/stretchr/testify/assert"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
)

func GetV1TestCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo-v1",
			Labels: map[string]string{
				k8sconst.CustomResourceDefinitionSchemaVersionKey: k8sconst.CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{},
					},
				},
			},
		},
	}
}

const labelkey = k8sconst.CustomResourceDefinitionSchemaVersionKey

var minVersion = versioncheck.MustVersion(k8sconst.CustomResourceDefinitionSchemaVersion)

func Test_CreateUpdateCRD(t *testing.T) {
	tests := []struct {
		name    string
		test    func() error
		wantErr bool
	}{{
		name: "install v1 crd on v1 apiserver",
		test: func() error {
			v1crd := GetV1TestCRD()
			client := fake.NewSimpleClientset()
			fakePoller := newFakePoller()
			return CreateUpdateCRD(client, v1crd, fakePoller, labelkey, minVersion)

		},
		wantErr: false,
	}}

	for _, tt := range tests {
		t.Log(tt.name)
		err := tt.test()
		assert.Equal(t, tt.wantErr, err != nil)
	}

}

func newFakePoller() fakePoller {
	return fakePoller{}
}

type fakePoller struct{}

func (f fakePoller) Poll(interval, duration time.Duration,
	conditionFn func() (bool, error)) error {
	return nil
}
