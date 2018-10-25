package prometheus_alerts

import (
	"code.cloudfoundry.org/indicators/pkg/indicator"
	"fmt"
)

type Rule struct {
	Alert       string
	Expr        string
	Labels      map[string]string
	Annotations map[string]string
}

type Document struct {
	Groups []Group
}

type Group struct {
	Name  string
	Rules []Rule
}

func AlertDocumentFrom(document indicator.Document) Document {
	rules := make([]Rule, 0)

	for _, ind := range document.Indicators {
		for _, threshold := range ind.Thresholds {
			rules = append(rules, ruleFrom(document, ind, threshold))
		}
	}

	return Document{
		Groups: []Group{{
			Name:  document.Product,
			Rules: rules,
		}},
	}
}

func ruleFrom(document indicator.Document, indicator indicator.Indicator, threshold indicator.Threshold) Rule {
	labels := map[string]string{
		"product": document.Product,
		"version": document.Version,
		"level":   threshold.Level,
	}

	if indicator.SLO != 0 {
		labels["slo"] = fmt.Sprintf("%+v", indicator.SLO)
	}

	for k, v := range document.Metadata {
		labels[k] = v
	}

	return Rule{
		Alert:       indicator.Name,
		Expr:        fmt.Sprintf("%s %s %+v", indicator.PromQL, threshold.GetComparator(), threshold.Value),
		Labels:      labels,
		Annotations: indicator.Documentation,
	}
}
