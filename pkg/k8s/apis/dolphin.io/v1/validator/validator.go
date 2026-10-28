package validator

import (
	"encoding/json"
	"fmt"

	apiextensionsinternal "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	crdembed "github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io/client/crd"
)

// DECValidator validates DolphinEnvoyConfig objects against their generated CRD schema.
type DECValidator struct {
	validator validation.SchemaCreateValidator
}

// NewDECValidator constructs a validator from the generated DolphinEnvoyConfig CRD.
func NewDECValidator() (*DECValidator, error) {
	crd := apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(crdembed.DolphinEnvoyConfigs, &crd); err != nil {
		return nil, fmt.Errorf("unmarshal generated DolphinEnvoyConfig CRD: %w", err)
	}

	var crv *apiextensionsv1.CustomResourceValidation
	for i := range crd.Spec.Versions {
		if crd.Spec.Versions[i].Name == "v1" {
			crv = crd.Spec.Versions[i].Schema
			break
		}
	}
	if crv == nil {
		return nil, fmt.Errorf("generated DolphinEnvoyConfig CRD has no v1 schema")
	}

	// Match the Kubernetes API server marshaling defaults applied before validation.
	b, err := json.Marshal(crv)
	if err != nil {
		return nil, fmt.Errorf("marshal DolphinEnvoyConfig CRD schema: %w", err)
	}
	var normalized apiextensionsv1.CustomResourceValidation
	if err := json.Unmarshal(b, &normalized); err != nil {
		return nil, fmt.Errorf("normalize DolphinEnvoyConfig CRD schema: %w", err)
	}

	var internal apiextensionsinternal.CustomResourceValidation
	if err := apiextensionsv1.Convert_v1_CustomResourceValidation_To_apiextensions_CustomResourceValidation(
		&normalized,
		&internal,
		nil,
	); err != nil {
		return nil, fmt.Errorf("convert DolphinEnvoyConfig CRD schema: %w", err)
	}

	schemaValidator, _, err := validation.NewSchemaValidator(internal.OpenAPIV3Schema)
	if err != nil {
		return nil, fmt.Errorf("create DolphinEnvoyConfig schema validator: %w", err)
	}

	return &DECValidator{validator: schemaValidator}, nil
}

// ValidateDEC validates a DolphinEnvoyConfig represented as an unstructured object.
func (v *DECValidator) ValidateDEC(dec *unstructured.Unstructured) error {
	if errs := validation.ValidateCustomResource(nil, dec.Object, v.validator); len(errs) > 0 {
		return errs.ToAggregate()
	}
	return nil
}

// DecodeDEC converts DEC YAML into an unstructured Kubernetes object.
func DecodeDEC(decYAML []byte) (*unstructured.Unstructured, error) {
	jsonBytes, err := yaml.YAMLToJSON(decYAML)
	if err != nil {
		return nil, fmt.Errorf("convert DEC YAML to JSON: %w", err)
	}

	dec := &unstructured.Unstructured{}
	if err := json.Unmarshal(jsonBytes, dec); err != nil {
		return nil, fmt.Errorf("decode DEC JSON: %w", err)
	}
	return dec, nil
}
