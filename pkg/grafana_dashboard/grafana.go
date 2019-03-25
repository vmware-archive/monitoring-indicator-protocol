package grafana_dashboard

import (
	"crypto/sha1"
	"fmt"
	"log"
	"strings"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func DashboardFilename(documentBytes []byte, productName string) string {
	return fmt.Sprintf("%s_%x.json", productName, sha1.Sum(documentBytes))
}

func DocumentToDashboard(document indicator.Document) GrafanaDashboard {
	return toGrafanaDashboard(document)
}

func toGrafanaDashboard(d indicator.Document) GrafanaDashboard {
	return GrafanaDashboard{
		Title:       getDashboardTitle(d),
		Rows:        toGrafanaRows(d.Layout.Sections),
		Annotations: toGrafanaAnnotations(d.Product, d.Metadata),
	}
}

func toGrafanaAnnotations(product indicator.Product, metadata map[string]string) GrafanaAnnotations {
	return GrafanaAnnotations{
		List: []GrafanaAnnotation{
			{
				Enable:      true,
				Expr:        fmt.Sprintf("ALERTS{product=\"%s\"%s}", product.Name, metadataToLabelSelector(metadata)),
				TagKeys:     "level",
				TitleFormat: "{{alertname}} is {{alertstate}} in the {{level}} threshold",
				IconColor:   "#1f78c1",
			},
		},
	}
}

func metadataToLabelSelector(metadata map[string]string) interface{} {
	var selector string
	for k, v := range metadata {
		selector = fmt.Sprintf("%s,%s=\"%s\"", selector, k, v)
	}
	return selector
}

func getDashboardTitle(d indicator.Document) string {
	if d.Layout.Title == "" {
		return fmt.Sprintf("%s - %s", d.Product.Name, d.Product.Version)
	}
	return d.Layout.Title
}

func toGrafanaRows(sections []indicator.Section) []GrafanaRow {
	var rows []GrafanaRow

	for _, i := range sections {
		rows = append(rows, sectionToGrafanaRow(i))
	}

	return rows
}

func getIndicatorTitle(i indicator.Indicator) string {
	title := i.Name
	if t, ok := i.Documentation["title"]; ok {
		title = t
	}
	return title
}

func toGrafanaPanel(i indicator.Indicator, title string) GrafanaPanel {
	return GrafanaPanel{
		Title: title,
		Type:  "graph",
		Targets: []GrafanaTarget{{
			Expression: strings.Replace(i.PromQL, "$step", "$__interval", -1),
		}},
		Thresholds: toGrafanaThresholds(i.Thresholds),
	}
}

func sectionToGrafanaRow(s indicator.Section) GrafanaRow {
	title := s.Title

	var panels []GrafanaPanel

	for _, s := range s.Indicators {
		panels = append(panels, toGrafanaPanel(s, getIndicatorTitle(s)))
	}

	return GrafanaRow{
		Title:  title,
		Panels: panels,
	}
}

func toGrafanaThresholds(thresholds []indicator.Threshold) []GrafanaThreshold {
	var grafanaThresholds []GrafanaThreshold
	for _, t := range thresholds {
		var comparator string
		switch {
		case t.Operator <= indicator.LessThanOrEqualTo:
			comparator = "lt"
		case t.Operator >= indicator.GreaterThanOrEqualTo:
			comparator = "gt"
		default:
			log.Printf(
				"grafana dashboards only support lt/gt thresholds, threshold skipped: %s: %s %v\n",
				t.Level,
				t.GetComparatorAbbrev(),
				t.Value,
			)
			continue
		}

		grafanaThresholds = append(grafanaThresholds, GrafanaThreshold{
			Value:     t.Value,
			ColorMode: t.Level,
			Op:        comparator,
			Fill:      true,
			Line:      true,
			Yaxis:     "left",
		})
	}

	return grafanaThresholds
}
