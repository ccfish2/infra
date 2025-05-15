package client

import (
	"os"
	"testing"

	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_GenerateCRDFromYaml(t *testing.T) {
	restConf := ctrl.GetConfigOrDie()

	if restConf == nil {
		panic("failed to retrieve config")
	}

	httpCli, err := rest.HTTPClientFor(restConf)
	if err != nil || httpCli == nil {
		panic("failed to create httpclient")
	}

	apiextCli, err := apiext.NewForConfigAndClient(restConf, httpCli)
	if err != nil {
		panic("failed to create apiextensions-apiserver client")
	}

	var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "k8s-test")
	CreateCustomResourceDefinitions(apiextCli, log.WithField("name", "k8sclient"))
}

func Test_GetPregeneratedCRD(t *testing.T) {

	var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "k8s-test")

	crds := AllDolphinCRDResourceNames(log.WithField("name", "k8s-cli-test"))
	assert.NotEqual(t, len(crds), 0)
	expect := CustomResourceDefinitionList()
	for crd := range crds {
		if _, ok := expect[crd]; !ok {
			t.Errorf(crd, " is not expected.")
		}
	}

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

func Test_constructV1CRD(t *testing.T) {
	var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "k8s-test")
	crds := AllDolphinCRDResourceNames(log.WithField("name", "k8s-cli-test"))
	assert.NotEqual(t, len(crds), 0)

	expect := CustomResourceDefinitionList()
	for crd := range crds {
		if _, ok := expect[crd]; !ok {
			t.Errorf(crd, " is not expected.")
		}
	}
<<<<<<< HEAD

	for _, _ = range crds {
		// dolphinapiextcrd, err := ConstructCRDFromYaml(yamlfile, log.WithField("k8s", "construct-crd-test"))
		// if err != nil {
		// 	t.Error(err)
		// }
		// constructV1CRD(crdMetaName, dolphinapiextcrd)
	}
=======
	// forget how the ConstructCRDFromYaml was originally wrote
	// for crdMetaName, yamlfile := range crds {
	// 	dolphinapiextcrd, err := ConstructCRDFromYaml(yamlfile, log.WithField("k8s", "construct-crd-test"))
	// 	if err != nil {
	// 		t.Error(err)
	// 	}
	// 	constructV1CRD(crdMetaName, dolphinapiextcrd)
	// }
>>>>>>> a6917ba (Make codecoverage work)
}
