package grafana_dashboard

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

func DashboardFilename(documentBytes []byte, productName string) string {
	return fmt.Sprintf("%s_%x.json", productName, sha1.Sum(documentBytes))
}

func DocumentToDashboard(document v1.IndicatorDocument) (*GrafanaDashboard, error) {
	return toGrafanaDashboard(document)
}

func toGrafanaDashboard(d v1.IndicatorDocument) (*GrafanaDashboard, error) {
	rows, err := toGrafanaRows(d)
	if err != nil {
		return nil, err
	}
	return &GrafanaDashboard{
		Title:       getDashboardTitle(d),
		Rows:        rows,
		Annotations: toGrafanaAnnotations(d.Spec.Product, d.ObjectMeta.Labels),
	}, nil
}

func toGrafanaAnnotations(product v1.Product, metadata map[string]string) GrafanaAnnotations {
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

func getDashboardTitle(d v1.IndicatorDocument) string {
	if d.Spec.Layout.Title == "" {
		return fmt.Sprintf("%s - %s", d.Spec.Product.Name, d.Spec.Product.Version)
	}
	return d.Spec.Layout.Title
}

func toGrafanaRows(document v1.IndicatorDocument) ([]GrafanaRow, error) {
	var rows []GrafanaRow

	for _, i := range document.Spec.Layout.Sections {
		row, err := sectionToGrafanaRow(i, document)
		if err != nil {
			return nil, err
		}
		rows = append(rows, *row)
	}

	return rows, nil
}

func getIndicatorTitle(i v1.IndicatorSpec) string {
	if t, ok := i.Documentation["title"]; ok {
		return t
	}

	return i.Name
}

func toGrafanaPanel(i v1.IndicatorSpec, title string) GrafanaPanel {
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

func sectionToGrafanaRow(section v1.Section, document v1.IndicatorDocument) (*GrafanaRow, error) {
	title := section.Title

	var panels []GrafanaPanel

	for _, indicatorName := range section.Indicators {
		i := document.Indicator(indicatorName)
		if i == nil {
			return nil, errors.New("indicator not found")
		}
		panels = append(panels, toGrafanaPanel(*i, getIndicatorTitle(*i)))
	}

	return &GrafanaRow{
		Title:  title,
		Panels: panels,
	}, nil
}

func toGrafanaThresholds(thresholds []v1.Threshold) []GrafanaThreshold {
	var grafanaThresholds []GrafanaThreshold
	for _, t := range thresholds {
		var comparator string
		switch {
		case t.Operator == v1.LessThanOrEqualTo || t.Operator == v1.LessThan:
			comparator = "lt"
		case t.Operator == v1.GreaterThanOrEqualTo || t.Operator == v1.GreaterThan:
			comparator = "gt"
		default:
			log.Printf("grafana dashboards only support lt/gt thresholds, threshold skipped: %v\n", t)
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
