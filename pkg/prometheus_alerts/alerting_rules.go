package prometheus_alerts

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

type Rule struct {
	Alert       string
	Expr        string
	For         string
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

func AlertDocumentFilename(documentBytes []byte, productName string) string {
	return fmt.Sprintf("%s_%x.yml", productName, sha1.Sum(documentBytes))
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
			Name:  document.Product.Name,
			Rules: rules,
		}},
	}
}

func ruleFrom(document indicator.Document, indicator indicator.Indicator, threshold indicator.Threshold) Rule {
	labels := map[string]string{
		"product": document.Product.Name,
		"version": document.Product.Version,
		"level":   threshold.Level,
	}

	for k, v := range document.Metadata {
		labels[k] = v
	}

	interpolatedPromQl := strings.Replace(indicator.PromQL, "$step", indicator.Alert.Step, -1)

	return Rule{
		Alert:       indicator.Name,
		Expr:        fmt.Sprintf("%s %s %+v", interpolatedPromQl, threshold.GetComparator(), threshold.Value),
		For:         indicator.Alert.For,
		Labels:      labels,
		Annotations: indicator.Documentation,
	}
}
