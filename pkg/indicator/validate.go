package indicator

import (
	"fmt"
	"strings"
)

func Validate(document Document) []error {
	es := make([]error, 0)

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

		if len(i.MetricRefs) == 0 {
			es = append(es, fmt.Errorf("indicators[%d] must reference at least 1 metric", idx))
		}
	}

	for sidx, s := range document.Documentation.Sections {

		for idx, i := range s.IndicatorRefs {
			if _, ok := FindIndicator(i, document.Indicators); !ok {
				es = append(es, fmt.Errorf("documentation.sections[%d].indicators[%d] references non-existent indicator.title (%s)", sidx, idx, i))
			}
		}

		for idx, i := range s.MetricRefs {
			if _, ok := FindMetric(i, document.Metrics); !ok {
				es = append(es, fmt.Errorf("documentation.sections[%d].metrics[%d] references non-existent metric.title (%s)", sidx, idx, i))
			}
		}
	}

	return es
}
