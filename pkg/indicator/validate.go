package indicator

import (
	"fmt"
	"github.com/prometheus/prometheus/promql"
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

	if document.Product == "" {
		es = append(es, fmt.Errorf("product is required"))
	}

	if document.Version == "" {
		es = append(es, fmt.Errorf("version is required"))
	}

	for idx, i := range document.Indicators {
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
	}

	return es
}
