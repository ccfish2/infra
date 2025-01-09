package client

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ccfish2/infra/pkg/k8s/apis/crdhelpers"
	k8sconst "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io"
	k8sconstv1 "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io/v1"
	"github.com/ccfish2/infra/pkg/k8s/client"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/ccfish2/infra/pkg/versioncheck"
)

const (
	DEPCRDName  = k8sconstv1.DEPKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
	DEPSCRDName = k8sconstv1.DEPSKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
	DECRDName   = k8sconstv1.DECKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
	DIDCRDName  = k8sconstv1.DIDKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
)

func RegisterCRDs(clientset client.Clientset) error {
	if err := CreateCustomResourceDefinitions(clientset); err != nil {
		return fmt.Errorf("Unable to create custom resource definition: %w", err)
	}

	return nil
}

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "k8s")

type CRDList struct {
	Name     string
	FullName string
}

func CustomResourceDefinitionList() map[string]*CRDList {
	return map[string]*CRDList{
		CRDResourceName(k8sconstv1.DEPName): {
			Name:     DEPCRDName,
			FullName: k8sconstv1.DEPName,
		},

		CRDResourceName(k8sconstv1.DEPSName): {
			Name:     DEPSCRDName,
			FullName: k8sconstv1.DEPSName,
		},

		CRDResourceName(k8sconstv1.DECName): {
			Name:     DECRDName,
			FullName: k8sconstv1.DECName,
		},

		CRDResourceName(k8sconstv1.DIDName): {
			Name:     DIDCRDName,
			FullName: k8sconstv1.DIDName,
		},
	}
}

func createCRD(crdfilePath string, crdMetaName string) func(clientset apiextensionsclient.Interface) error {
	return func(clientset apiextensionsclient.Interface) error {

		dolphinCRD := apiextensionsv1.CustomResourceDefinition{}
		crdBytes, err := os.ReadFile(crdfilePath)
		if err != nil {
			panic("Failed to retrieve file ")

		}
		err = yaml.Unmarshal(crdBytes, &dolphinCRD)
		if err != nil {
			panic(err)
		}

		return crdhelpers.CreateUpdateCRD(
			clientset,
			constructV1CRD(crdMetaName, dolphinCRD),
			crdhelpers.NewDefaultPoller(),
			k8sconst.CustomResourceDefinitionSchemaVersionKey,
			versioncheck.MustVersion(k8sconst.CustomResourceDefinitionSchemaVersion),
		)
	}
}

func constructV1CRD(
	name string,
	template apiextensionsv1.CustomResourceDefinition,
) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconst.CustomResourceDefinitionSchemaVersionKey: k8sconst.CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: k8sconstv1.CustomResourceDefinitionGroup,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:       template.Spec.Names.Kind,
				Plural:     template.Spec.Names.Plural,
				ShortNames: template.Spec.Names.ShortNames,
				Singular:   template.Spec.Names.Singular,
			},
			Scope:    template.Spec.Scope,
			Versions: template.Spec.Versions,
		},
	}
}

func AllDolphinCRDResourceNames() map[string]string {
	curP := filepath.Base(".")
	crdBases := fmt.Sprintf("%s/crd/bases", curP)
	fSystem := os.DirFS(crdBases)
	ret := map[string]string{}
	fs.WalkDir(fSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path != "." {
			if strings.HasSuffix(path, ".yaml") {
				path = strings.ReplaceAll(path, ".yaml", "")
			}
			crdBases = fmt.Sprintf("%s/%s", crdBases, path)
			log.Info("yaml ", path, " ::file ", ret[path])
		}

		return nil
	})
	return ret
}

func CreateCustomResourceDefinitions(clientset apiextensionsclient.Interface) error {
	g, _ := errgroup.WithContext(context.Background())

	crds := CustomResourceDefinitionList()

	for r, filepath := range AllDolphinCRDResourceNames() {
		if crd, ok := crds[r]; ok {
			g.Go(func() error {
				return createCRD(filepath, crd.FullName)(clientset)
			})
		} else {
			log.Fatalf("Unknown resource %s. Please update pkg/k8s/apis/dolphin.io/client to understand this type.", r)
		}
	}

	return g.Wait()
}

func agentCRDResourceNames() []string {
	result := []string{
		CRDResourceName(k8sconstv1.DEPName),
	}
	return result
}

func CRDResourceName(crd string) string {
	return "crd:" + crd
}

func AgentCRDResourceNames() []string {
	return agentCRDResourceNames()
}
