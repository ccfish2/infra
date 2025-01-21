package client

import (
	"os"
	"testing"

	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func Test_GetPregeneratedCRD(t *testing.T) {
	var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "k8s-test")
	crds := AllDolphinCRDResourceNames(log.WithField("name", "k8s-cli-test"))
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
