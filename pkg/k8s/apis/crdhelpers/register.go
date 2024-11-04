package crdhelpers

import (
	"context"
	goerrors "errors"
	"fmt"
	"time"

	semver "github.com/blang/semver/v4"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// subsysK8s is the value for logfields.LogSubsys
const subsysK8s = "k8s"

func CreateUpdateCRD(
	clientset apiextensionsclient.Interface,
	crd *apiextensionsv1.CustomResourceDefinition,
	poller poller,
	crdSchemaVersionLabelKey string,
	minCRDSchemaVersion semver.Version,
) error {

	v1CRDClient := clientset.ApiextensionsV1()
	clusterCRD, err := v1CRDClient.CustomResourceDefinitions().Get(
		context.TODO(),
		crd.ObjectMeta.Name,
		metav1.GetOptions{})
	if errors.IsNotFound(err) {

		clusterCRD, err = v1CRDClient.CustomResourceDefinitions().Create(
			context.TODO(),
			crd,
			metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			return nil
		}
	}
	if err != nil {
		return err
	}
	if err := updateV1CRD(crd, clusterCRD, v1CRDClient, poller, crdSchemaVersionLabelKey, minCRDSchemaVersion); err != nil {
		return err
	}
	if err := waitForV1CRD(clusterCRD, v1CRDClient, poller); err != nil {
		return err
	}

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
	if !ok {
		return true
	}

	return false
}

func updateV1CRD(
	crd, clusterCRD *apiextensionsv1.CustomResourceDefinition,
	client v1client.CustomResourceDefinitionsGetter,
	poller poller,
	crdSchemaVersionLabelKey string,
	minCRDSchemaVersion semver.Version,
) error {
	if crd.Spec.Versions[0].Schema != nil && needsUpdateV1(clusterCRD, crdSchemaVersionLabelKey, minCRDSchemaVersion) {
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

	return nil
}

func waitForV1CRD(
	crd *apiextensionsv1.CustomResourceDefinition,
	client v1client.CustomResourceDefinitionsGetter,
	poller poller,
) error {
	fmt.Printf("Waiting for CRD (CustomResourceDefinition) to be available...")

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
					fmt.Printf("Name conflict for CRD")
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
		return fmt.Errorf("error occurred waiting for CRD: %w", err)
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
