package indicator

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/promql"
)

func Validate(document Document) []error {
	es := make([]error, 0)
	if document.APIVersion == "" {
		es = append(es, fmt.Errorf("apiVersion is required"))
	}

	if document.APIVersion != "v0" {
		es = append(es, fmt.Errorf("only apiVersion v0 is supported"))
	}

	if document.Product.Name == "" {
		es = append(es, fmt.Errorf("product name is required"))
	}

	if document.Product.Version == "" {
		es = append(es, fmt.Errorf("product version is required"))
	}

	for k := range document.Metadata {
		if k == "step" {
			es = append(es, fmt.Errorf("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"))
		}
	}

	for idx, i := range document.Indicators {
		es = validateIndicator(i, es, idx)
	}

	return es
}

func validateIndicator(i Indicator, es []error, idx int) []error {
	if strings.TrimSpace(i.Name) == "" {
		es = append(es, fmt.Errorf("indicators[%d] name is required", idx))
	}
	labels, err := promql.ParseMetric(i.Name)
	if err != nil || labels.Len() > 1 {
		es = append(es, fmt.Errorf("indicators[%d] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)", idx))
	}
	if strings.TrimSpace(i.PromQL) == "" {
		es = append(es, fmt.Errorf("indicators[%d] promql is required", idx))
	}
	for tdx, threshold := range i.Thresholds {
		if threshold.Operator == Undefined {
			es = append(es, fmt.Errorf("indicators[%d].thresholds[%d] value is required, one of [lt, lte, eq, neq, gte, gt] must be provided as a float", idx, tdx))
		}
	}

	es = validateChartType(i.Presentation.ChartType, es, idx)

	return es
}
func validateChartType(chartType ChartType, es []error, idx int) []error {
	valid := false
	for _, validChartType := range ChartTypes {
		if chartType == validChartType {
			valid = true
		}
	}
	if !valid {
		es = append(es, fmt.Errorf("indicators[%d] invalid chartType provided: '%s' - valid chart types are %v", idx, chartType, ChartTypes))
	}

	return es
}
