package grafana_dashboard

import (
	"fmt"
	"log"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func DocumentToDashboard(document indicator.Document) GrafanaDashboard {
	return toGrafanaDashboard(document)
}

func toGrafanaDashboard(d indicator.Document) GrafanaDashboard {
	return GrafanaDashboard{
		Title: getDashboardTitle(d),
		Rows:  toGrafanaRows(d.Indicators, d.Layout.Sections),
	}
}

func getDashboardTitle(d indicator.Document) string {
	if d.Layout.Title == "" {
		return fmt.Sprintf("%s - %s", d.Product.Name, d.Product.Version)
	}
	return d.Layout.Title
}

func toGrafanaRows(indicators []indicator.Indicator, sections []indicator.Section) []GrafanaRow {
	var rows []GrafanaRow
	if sections == nil {
		for _, i := range indicators {
			rows = append(rows, toGrafanaRow(i))
		}
	} else {
		for _, i := range sections {
			rows = append(rows, sectionToGrafanaRow(i))
		}
	}

	return rows
}

func toGrafanaRow(i indicator.Indicator) GrafanaRow {
	title := GetIndicatorTitle(i)

	return GrafanaRow{
		Title: title,
		Panels: []GrafanaPanel{ToGrafanaPanel(i, title)},
	}
}

func GetIndicatorTitle(i indicator.Indicator) string {
	title := i.Name
	if t, ok := i.Documentation["title"]; ok {
		title = t
	}
	return title
}

func ToGrafanaPanel(i indicator.Indicator, title string) GrafanaPanel {
	return GrafanaPanel{
		Title: title,
		Type:  "graph",
		Targets: []GrafanaTarget{{
			Expression: i.PromQL,
		}},
		Thresholds: toGrafanaThresholds(i.Thresholds),
	}
}

func sectionToGrafanaRow(s indicator.Section) GrafanaRow {
	title := s.Title

	var panels []GrafanaPanel

	for _, s := range s.Indicators {
		panels = append(panels, ToGrafanaPanel(s, GetIndicatorTitle(s)))
	}

	return GrafanaRow{
		Title: title,
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
