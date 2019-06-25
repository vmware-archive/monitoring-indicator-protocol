package indicator

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/promql"
)

func (document *Document) Validate(apiVersion string) []error {
	es := make([]error, 0)
	if document.APIVersion == "" {
		es = append(es, fmt.Errorf("apiVersion is required"))
	}

	if document.Product.Name == "" {
		es = append(es, fmt.Errorf("product name is required"))
	}

	if document.Product.Version == "" {
		es = append(es, fmt.Errorf("product version is required"))
	}

	for k := range document.Metadata {
		if strings.ToLower(k) == "step" {
			es = append(es, fmt.Errorf("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"))
		}
	}

	for idx, i := range document.Indicators {
		es = append(es, i.Validate(idx)...)
	}

	for sectionIdx, section := range document.Layout.Sections {
		for idx, i := range section.Indicators {
			if _, found := document.GetIndicator(i); !found {
				es = append(es, fmt.Errorf("layout sections[%d] indicators[%d] references a non-existent indicator: %s", sectionIdx, idx, i))
			}
		}
	}

	if document.APIVersion != apiVersion {
		es = append(es, fmt.Errorf("only apiVersion %s is supported", apiVersion))
	}

	return es
}

func (indicator *Indicator) Validate(idx int) []error {
	var es []error
	if strings.TrimSpace(indicator.Name) == "" {
		es = append(es, fmt.Errorf("indicators[%d] name is required", idx))
	}
	labels, err := promql.ParseMetric(indicator.Name)
	if err != nil || labels.Len() > 1 {
		es = append(es, fmt.Errorf("indicators[%d] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)", idx))
	}
	if strings.TrimSpace(indicator.PromQL) == "" {
		es = append(es, fmt.Errorf("indicators[%d] promql is required", idx))
	}
	for tdx, threshold := range indicator.Thresholds {
		if threshold.Operator == Undefined {
			es = append(es, fmt.Errorf("indicators[%d].thresholds[%d] value is required, one of [lt, lte, eq, neq, gte, gt] must be provided as a float", idx, tdx))
		}
	}

	es = append(es, indicator.Presentation.ChartType.Validate(idx)...)

	return es
}

func (chartType *ChartType) Validate(idx int) []error {
	var es []error
	valid := false
	for _, validChartType := range ChartTypes {
		if *chartType == validChartType {
			valid = true
		}
	}
	if !valid {
		es = append(es, fmt.Errorf("indicators[%d] invalid chartType provided: '%s' - valid chart types are %v", idx, *chartType, ChartTypes))
	}

	return es
}
