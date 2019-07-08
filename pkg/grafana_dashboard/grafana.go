package grafana_dashboard

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func DashboardFilename(documentBytes []byte, productName string) string {
	return fmt.Sprintf("%s_%x.json", productName, sha1.Sum(documentBytes))
}

func DocumentToDashboard(document indicator.Document) (*GrafanaDashboard, error) {
	return toGrafanaDashboard(document)
}

func toGrafanaDashboard(d indicator.Document) (*GrafanaDashboard, error) {
	rows, err := toGrafanaRows(d)
	if err != nil {
		return nil, err
	}
	return &GrafanaDashboard{
		Title:       getDashboardTitle(d),
		Rows:        rows,
		Annotations: toGrafanaAnnotations(d.Product, d.Metadata),
	}, nil
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

func toGrafanaRows(document indicator.Document) ([]GrafanaRow, error) {
	var rows []GrafanaRow

	for _, i := range document.Layout.Sections {
		row, err := sectionToGrafanaRow(i, document)
		if err != nil {
			return nil, err
		}
		rows = append(rows, *row)
	}

	return rows, nil
}

func getIndicatorTitle(i indicator.Indicator) string {
	if t, ok := i.Documentation["title"]; ok {
		return t
	}

	return i.Name
}

func toGrafanaPanel(i indicator.Indicator, title string) GrafanaPanel {
	replacementString := replaceStep(i.PromQL)
	return GrafanaPanel{
		Title: title,
		Type:  "graph",
		Targets: []GrafanaTarget{{
			Expression: replacementString,
		}},
		Thresholds: toGrafanaThresholds(i.Thresholds),
	}
}

func replaceStep(str string) string {
	reg := regexp.MustCompile(`(?i)\$step\b`)
	return reg.ReplaceAllString(str, `$$__interval`)
}

func sectionToGrafanaRow(section indicator.Section, document indicator.Document) (*GrafanaRow, error) {
	title := section.Title

	var panels []GrafanaPanel

	for _, indicatorName := range section.Indicators {
		i, found := document.GetIndicator(indicatorName)
		if !found {
			return nil, errors.New("indicator not found")
		}
		panels = append(panels, toGrafanaPanel(i, getIndicatorTitle(i)))
	}

	return &GrafanaRow{
		Title:  title,
		Panels: panels,
	}, nil
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
