package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func Test_GetPregeneratedCRD(t *testing.T) {
	crds := AllDolphinCRDResourceNames()
	fmt.Println(crds)
	assert.NotEqual(t, len(crds), 0)

	for crdMetaName, crdfilePath := range crds {
		dolphinCRD := apiextensionsv1.CustomResourceDefinition{}
		crdBytes, err := os.ReadFile(crdfilePath)
		if err != nil {
			log.Error("Failed to retrieve file ")

		}
		err = yaml.Unmarshal(crdBytes, &dolphinCRD)
		if err != nil {
			log.Error("Error unmarshalling pregenerated CRD, ", err)
		}
		constructV1CRD(crdMetaName, dolphinCRD)
	}
}
