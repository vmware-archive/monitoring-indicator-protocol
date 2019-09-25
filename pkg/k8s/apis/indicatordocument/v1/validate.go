package v1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/prometheus/prometheus/promql"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/asset"
)

func (doc *IndicatorDocument) Validate(supportedApiVersion ...string) []error {
	es := make([]error, 0)

	// Instead of duplicated validation code, we can just marshal the document and then validate its bytes
	docBytes, err := yaml.Marshal(doc)
	if err != nil {
		es = append(es, err)
	}

	errs, valid := ValidateBytesBySchema(docBytes, "IndicatorDocument")
	if !valid {
		es = append(es, errs...)
	}

	for k := range doc.ObjectMeta.Labels {
		if strings.ToLower(k) == "step" {
			es = append(es, errors.New("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"))
		}
	}

	for idx, i := range doc.Spec.Indicators {
		es = append(es, i.Validate(idx, doc.APIVersion)...)
	}

	for sectionIdx, section := range doc.Spec.Layout.Sections {
		for idx, indicatorName := range section.Indicators {
			if indicator := doc.Indicator(indicatorName); indicator == nil {
				es = append(es, fmt.Errorf("layout sections[%d] indicators[%d] references a non-existent indicator", sectionIdx, idx))
			}
		}
	}

	return es
}

func (is *IndicatorSpec) Validate(indicatorIndex int, apiVersion string) []error {
	var es []error

	indicatorBytes, err := yaml.Marshal(is)
	if err != nil {
		es = append(es, err)
	}

	errs, valid := ValidateBytesBySchema(indicatorBytes, "IndicatorSpec")
	for i, err := range errs {
		errs[i] = fmt.Errorf("indicators[%d] is invalid by schema: %s", indicatorIndex, err)
	}

	if !valid {
		es = append(es, errs...)
	}

	labels, err := promql.ParseMetric(is.Name)
	if err != nil || labels.Len() > 1 {
		es = append(es, fmt.Errorf("indicators[%d] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)", indicatorIndex))
	}

	return es
}

// Validates provided YAML is in correct v1 format by OpenAPI Schema
func ValidateBytesBySchema(docBytes []byte, schemaName string) ([]error, bool) {
	schemaBytes, err := asset.Asset("schemas.yml")
	if err != nil {
		return []error{err}, false
	}

	var schemaHolder struct {
		IndicatorDocumentSchema spec.Schema `json:"IndicatorDocument"`
		IndicatorSchema         spec.Schema `json:"IndicatorSpec"`
	}
	var rootSchema interface{}
	err = yaml.Unmarshal(schemaBytes, &rootSchema)
	if err != nil {
		return []error{err}, false
	}

	err = yaml.Unmarshal(schemaBytes, &schemaHolder)
	if err != nil {
		return []error{err}, false
	}

	var schema spec.Schema
	switch schemaName {
	case "IndicatorDocument":
		schema = schemaHolder.IndicatorDocumentSchema
	case "IndicatorSpec":
		schema = schemaHolder.IndicatorSchema
	default:
		return []error{fmt.Errorf("invalid schema name '%s'", schemaName)}, false
	}
	validator := validate.NewSchemaValidator(&schema, rootSchema, schemaName, strfmt.Default)

	var document interface{}
	err = yaml.Unmarshal(docBytes, &document)
	if err != nil {
		return []error{err}, false
	}

	errs := make([]error, 0)
	for _, e := range validator.Validate(document).Errors {
		errs = append(errs, errors.New(e.Error()))
	}

	return errs, validator.Validate(document).IsValid()
}
