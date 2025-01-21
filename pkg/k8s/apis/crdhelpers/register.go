package crdhelpers

import (
	"context"
	goerrors "errors"
	"time"

	semver "github.com/blang/semver/v4"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
	"github.com/sirupsen/logrus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// subsysK8s is the value for logfields.LogSubsys
const subsysK8s = "k8s"

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, subsysK8s)

func CreateUpdateCRD(
	clientset apiextensionsclient.Interface,
	crd *apiextensionsv1.CustomResourceDefinition,
	poller poller,
	crdSchemaVersionLabelKey string,
	minCRDSchemaVersion semver.Version,
) error {
	scopedlog := log.WithField(" name ", crd.Name)
	v1CRDClient := clientset.ApiextensionsV1()
	clusterCRD, err := v1CRDClient.CustomResourceDefinitions().Get(
		context.TODO(),
		crd.ObjectMeta.Name,
		metav1.GetOptions{})
	if errors.IsNotFound(err) {
		scopedlog.Info("creating CRD ....")
		clusterCRD, err = v1CRDClient.CustomResourceDefinitions().Create(
			context.TODO(),
			crd,
			metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			return nil
		}
	}

	if err != nil {
		scopedlog.Error("Failed to create CRD ", err)
		return err
	}

	if err := updateV1CRD(scopedlog, crd, clusterCRD, v1CRDClient, poller, crdSchemaVersionLabelKey, minCRDSchemaVersion); err != nil {
		scopedlog.Error("Failed to update CRD ", err)
		return err
	}

	if err := waitForV1CRD(scopedlog, clusterCRD, v1CRDClient, poller); err != nil {
		scopedlog.Error("Could not got CRD successfully ", err)
		return err
	}

	scopedlog.Info("CRD is installed successfully.")
	return nil
}

func needsUpdateV1(
	clusterCRD *apiextensionsv1.CustomResourceDefinition,
	crdSchemaVersionLabelKey string,
	minCRDSchemaVersion semver.Version,
) bool {
	if clusterCRD.Spec.Versions[0].Schema == nil {
		return true
	}
	_, ok := clusterCRD.Labels[crdSchemaVersionLabelKey]
	if ok {
		// no schema version detected
		return true
	}

	return false
}

func updateV1CRD(
	scopedLog *logrus.Entry,
	crd, clusterCRD *apiextensionsv1.CustomResourceDefinition,
	client v1client.CustomResourceDefinitionsGetter,
	poller poller,
	crdSchemaVersionLabelKey string,
	minCRDSchemaVersion semver.Version,
) error {
	scopedLog.Info("Checking if CRD needs updating...")
	if crd.Spec.Versions[0].Schema != nil && needsUpdateV1(clusterCRD, crdSchemaVersionLabelKey, minCRDSchemaVersion) {
		scopedLog.Info(" Updating CRD ...")
		err := poller.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
			var err error
			clusterCRD, err = client.CustomResourceDefinitions().Get(
				context.TODO(),
				crd.ObjectMeta.Name,
				metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			if needsUpdateV1(clusterCRD, crdSchemaVersionLabelKey, minCRDSchemaVersion) {
				scopedLog.Info("CRD validation is different, installing CRD ...")
				clusterCRD.ObjectMeta.Labels = crd.ObjectMeta.Labels
				clusterCRD.Spec = crd.Spec

				clusterCRD.Spec.PreserveUnknownFields = false

				_, err := client.CustomResourceDefinitions().Update(
					context.TODO(),
					clusterCRD,
					metav1.UpdateOptions{})
				switch {
				case errors.IsConflict(err): // Occurs as Operators race to update CRDs.
					return false, nil
				case err == nil:
					scopedLog.Info("Updating CRD Validation successfully !")
					return true, nil
				}

				return false, err
			}

			return true, nil
		})
		if err != nil {
			return err
		}
	}
	scopedLog.Info("Updating CRD successfully. ")
	return nil
}

func waitForV1CRD(
	scopedLog *logrus.Entry,
	crd *apiextensionsv1.CustomResourceDefinition,
	client v1client.CustomResourceDefinitionsGetter,
	poller poller,
) error {
	log.Debug("Waiting for CRD (CustomResourceDefinition) to be available...")

	err := poller.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1.Established:
				if cond.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}
			case apiextensionsv1.NamesAccepted:
				if cond.Status == apiextensionsv1.ConditionFalse {
					err := goerrors.New(cond.Reason)
					log.WithError(err).Error("Name conflict for CRD")
					return false, err
				}
			}
		}

		var err error
		if crd, err = client.CustomResourceDefinitions().Get(
			context.TODO(),
			crd.ObjectMeta.Name,
			metav1.GetOptions{}); err != nil {
			return false, err
		}
		return false, err
	})
	if err != nil {
		scopedLog.Error("error occurred waiting for CRD: ", err)
		return err
	}

	return nil
}

type poller interface {
	Poll(interval, duration time.Duration, conditionFn func() (bool, error)) error
}

func NewDefaultPoller() defaultPoll {
	return defaultPoll{}
}

type defaultPoll struct{}

func (p defaultPoll) Poll(
	interval, duration time.Duration,
	conditionFn func() (bool, error),
) error {
	return wait.Poll(interval, duration, conditionFn)
}
