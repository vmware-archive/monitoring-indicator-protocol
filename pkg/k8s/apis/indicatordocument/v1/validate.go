package v1

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/prometheus/prometheus/promql"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
)

// TODO This can be simplified once we remove v0
func (id *IndicatorDocument) Validate(supportedApiVersion ...string) []error {
	es := make([]error, 0)
	if id.APIVersion == "" {
		es = append(es, errors.New("apiVersion is required"))
	}

	if id.Kind != "IndicatorDocument" {
		es = append(es, errors.New("`kind` must be \"IndicatorDocument\""))
	}

	if reflect.DeepEqual(id.Spec, IndicatorDocumentSpec{}) {
		es = append(es, errors.New("spec is required"))
	}

	if id.Spec.Product.Name == "" {
		es = append(es, errors.New("product name is required"))
	}

	if id.Spec.Product.Version == "" {
		es = append(es, errors.New("product version is required"))
	}

	for k := range id.ObjectMeta.Labels {
		if strings.ToLower(k) == "step" {
			es = append(es, errors.New("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"))
		}
	}

	for idx, i := range id.Spec.Indicators {
		es = append(es, i.Validate(idx, id.APIVersion)...)
	}

	for sectionIdx, section := range id.Spec.Layout.Sections {
		for idx, indicatorName := range section.Indicators {
			if indicator := id.Indicator(indicatorName); indicator == nil {
				es = append(es, fmt.Errorf("layout sections[%d] indicators[%d] references a non-existent indicator", sectionIdx, idx))
			}
		}
	}

	apiVersionValid := false
	for _, version := range supportedApiVersion {
		if id.APIVersion == version {
			apiVersionValid = true
		}
	}
	if !apiVersionValid {
		es = append(es, fmt.Errorf("invalid apiVersion, supported versions are: %v", supportedApiVersion))
	}

	return es
}

func (is *IndicatorSpec) Validate(idx int, apiVersion string) []error {
	var es []error
	if strings.TrimSpace(is.Name) == "" {
		es = append(es, fmt.Errorf("indicators[%d] name is required", idx))
	}
	labels, err := promql.ParseMetric(is.Name)
	if err != nil || labels.Len() > 1 {
		es = append(es, fmt.Errorf("indicators[%d] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)", idx))
	}
	if strings.TrimSpace(is.PromQL) == "" {
		es = append(es, fmt.Errorf("indicators[%d] promql is required", idx))
	}
	for tdx, threshold := range is.Thresholds {
		if threshold.Operator == UndefinedOperator && apiVersion == api_versions.V0 {
			es = append(es, fmt.Errorf("indicators[%d].thresholds[%d] value is required, one of [lt, lte, eq, neq, gte, gt] must be provided as a float", idx, tdx))
		} else if threshold.Operator == UndefinedOperator && apiVersion == api_versions.V1 {
			es = append(es, fmt.Errorf("indicators[%d].thresholds[%d] operator [lt, lte, eq, neq, gte, gt] is required", idx, tdx))
		}
	}

	if is.Type == UndefinedType {
		es = append(es, fmt.Errorf(
			"indicators[%d] given invalid type. Must be one of [sli, kpi, other] (if absent from the yaml, defaults to `other`)", idx))
	}

	es = append(es, is.Presentation.ChartType.Validate(idx)...)

	return es
}

// Validates provided YAML is in correct v1 format by OpenAPI Schema
func ValidateDocumentBytes(docBytes []byte) ([]error, bool) {
	schemaBytes, err := Asset("schemas.yml")
	if err != nil {
		return []error{err}, false
	}

	var schema struct {
		IndicatorDocumentSchema spec.Schema `json:"IndicatorDocument"`
	}
	var rootSchema interface{}
	err = yaml.Unmarshal(schemaBytes, &rootSchema)
	err = yaml.Unmarshal(schemaBytes, &schema)
	validator := validate.NewSchemaValidator(&schema.IndicatorDocumentSchema, rootSchema, "IndicatorDocument", strfmt.Default)

	var document interface{}
	err = yaml.Unmarshal(docBytes, &document)

	return validator.Validate(document).Errors, validator.Validate(document).IsValid()
}
