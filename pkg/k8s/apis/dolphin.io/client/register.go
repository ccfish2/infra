package client

import (
	"context"
	"fmt"
	"io/fs"
	"os"
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
	"github.com/sirupsen/logrus"
)

const (
	DEPCRDName  = k8sconstv1.DEPKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
	DEPSCRDName = k8sconstv1.DEPSKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
	DECRDName   = k8sconstv1.DECKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
	DIDCRDName  = k8sconstv1.DIDKindDefinition + "/" + k8sconstv1.CustomResourceDefinitionVersion
)

func RegisterCRDs(clientset client.Clientset, scopedlog *logrus.Entry) error {
	scopedlog.Info("RegisterCRDs ...")
	if err := CreateCustomResourceDefinitions(clientset, scopedlog); err != nil {
		return fmt.Errorf("unable to create custom resource definition: %w", err)
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

func createCRD(crdfilePath string, crdMetaName string, scopedlog *logrus.Entry) func(clientset apiextensionsclient.Interface) error {
	scopedlog.Info("creating CRD ", crdfilePath, " MetaName ", crdMetaName)
	return func(clientset apiextensionsclient.Interface) error {
		dolphinCRD := apiextensionsv1.CustomResourceDefinition{}
		crdBytes, err := os.ReadFile(crdfilePath)
		if err != nil {
			scopedlog.Error(" Failed to retrieve the CRD file ", crdfilePath)
			panic(err)
		}
		err = yaml.Unmarshal(crdBytes, &dolphinCRD)
		if err != nil {
			scopedlog.Error(" Failed to unmarshal crd file ", crdfilePath)
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

func AllDolphinCRDResourceNames(scopedlog *logrus.Entry) map[string]string {
	crdBases := "/tmp/dolphin"
	fSystem := os.DirFS(crdBases)
	ret := map[string]string{}
	fs.WalkDir(fSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path != "." {
			if strings.HasSuffix(path, ".yaml") {
				ret["crd:"+strings.ReplaceAll(path, ".yaml", "")] = fmt.Sprintf("%s/%s", crdBases, path)
			}
		}
		return nil
	})
	scopedlog.Info(ret)
	return ret
}

func CreateCustomResourceDefinitions(clientset apiextensionsclient.Interface, scopedlog *logrus.Entry) error {
	if clientset == nil {
		panic("client to access k8s api could not be nil.")
	}
	scopedlog.Info(" Create CRD ...")
	g, _ := errgroup.WithContext(context.Background())

	crds := CustomResourceDefinitionList()
	scopedlog.Info(" totally, we need creating #", len(crds), "  CRDs are ", crds)
	for r, filepath := range AllDolphinCRDResourceNames(scopedlog) {
		crd, ok := crds[r]
		if ok {
			scopedlog.Info(" Invoking creating CRD ", filepath, " itsname ", crd.FullName)
			g.Go(func() error {
				return createCRD(filepath, crd.FullName, scopedlog)(clientset)
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
