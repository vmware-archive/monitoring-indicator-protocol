package indicator

import (
	"fmt"
	"strings"
)

func Validate(document Document) []error {
	es := make([]error, 0)

	if document.APIVersion == "" {
		es = append(es, fmt.Errorf("apiVersion is required"))
	}

	if document.APIVersion != "v0" {
		es = append(es, fmt.Errorf("only apiVersion v0 is supported"))
	}

	if _, ok := document.Labels["product"]; !ok {
		es = append(es, fmt.Errorf("document labels.product is required"))
	}

	for idx, m := range document.Metrics {
		if strings.TrimSpace(m.Title) == "" {
			es = append(es, fmt.Errorf("metrics[%d] title is required", idx))
		}

		if strings.TrimSpace(m.Description) == "" {
			es = append(es, fmt.Errorf("metrics[%d] description is required", idx))
		}

		if strings.TrimSpace(m.Name) == "" {
			es = append(es, fmt.Errorf("metrics[%d] name is required", idx))
		}

		if strings.TrimSpace(m.SourceID) == "" {
			es = append(es, fmt.Errorf("metrics[%d] source_id is required", idx))
		}

		if strings.TrimSpace(m.Origin) == "" {
			es = append(es, fmt.Errorf("metrics[%d] origin is required", idx))
		}

		if strings.TrimSpace(m.Type) == "" {
			es = append(es, fmt.Errorf("metrics[%d] type is required", idx))
		}

		if strings.TrimSpace(m.Frequency) == "" {
			es = append(es, fmt.Errorf("metrics[%d] frequency is required", idx))
		}
	}

	for idx, i := range document.Indicators {
		if strings.TrimSpace(i.Name) == "" {
			es = append(es, fmt.Errorf("indicators[%d] name is required", idx))
		}

		if strings.TrimSpace(i.Title) == "" {
			es = append(es, fmt.Errorf("indicators[%d] title is required", idx))
		}

		if strings.TrimSpace(i.Description) == "" {
			es = append(es, fmt.Errorf("indicators[%d] description is required", idx))
		}

		if strings.TrimSpace(i.PromQL) == "" {
			es = append(es, fmt.Errorf("indicators[%d] promql is required", idx))
		}

		if strings.TrimSpace(i.Response) == "" {
			es = append(es, fmt.Errorf("indicators[%d] response is required", idx))
		}

		if strings.TrimSpace(i.Measurement) == "" {
			es = append(es, fmt.Errorf("indicators[%d] measurement is required", idx))
		}

		if len(i.Metrics) == 0 {
			es = append(es, fmt.Errorf("indicators[%d] must reference at least 1 metric", idx))
		}
	}

	return es
}
