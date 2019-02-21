package grafana_dashboard

import (
	"log"

	"github.com/pivotal/indicator-protocol/pkg/indicator"
)

func DocumentToDashboard(document indicator.Document) GrafanaDashboard {
	return toGrafanaDashboard(document)
}

func toGrafanaDashboard(d indicator.Document) GrafanaDashboard {
	return GrafanaDashboard{
		Title: d.Layout.Title,
		Rows:  toGrafanaRows(d.Indicators),
	}
}

func toGrafanaRows(indicators []indicator.Indicator) []GrafanaRow {
	var rows []GrafanaRow
	for _, i := range indicators {
		rows = append(rows, toGrafanaRow(i))
	}

	return rows
}

func toGrafanaRow(i indicator.Indicator) GrafanaRow {
	title := i.Name
	if t, ok := i.Documentation["title"]; ok {
		title = t
	}

	return GrafanaRow{
		Title: title,
		Panels: []GrafanaPanel{{
			Title: title,
			Type:  "graph",
			Targets: []GrafanaTarget{{
				Expression: i.PromQL,
			}},
			Thresholds: toGrafanaThresholds(i.Thresholds),
		}},
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
