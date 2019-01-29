package grafana_dashboard

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/pivotal/indicator-protocol/pkg/indicator"
)

var documatationTmpl = template.Must(template.New("grafana").Parse(grafanaTemplate))
var indicatorTmpl = template.Must(template.New("grafanaRow").Parse(grafanaIndicatorTemplate))

func DocumentToDashboard(document indicator.Document) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := documatationTmpl.Execute(buffer, grafanaDashboard{document})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type grafanaDashboard struct {
	document indicator.Document
}

func (g grafanaDashboard) Title() string {
	return g.document.Layout.Title
}

func (g grafanaDashboard) Indicators() string {
	grafanaIndicators := make([]string, len(g.document.Indicators))
	for idx, i := range g.document.Indicators {
		buffer := bytes.NewBuffer(nil)
		err := indicatorTmpl.Execute(buffer, grafanaIndicator{i})

		if err != nil {
			fmt.Fprint(buffer, "{}")
		}

		grafanaIndicators[idx] = buffer.String()
	}

	return fmt.Sprintf(`[%s]`, strings.Join(grafanaIndicators, ","))
}

type grafanaIndicator struct {
	indicator indicator.Indicator
}

func (i grafanaIndicator) Title() string {
	return i.indicator.Name
}

func (i grafanaIndicator) Promql() string {
	return strings.Replace(i.indicator.PromQL, `"`, `\"`, -1)
}

func (i grafanaIndicator) Thresholds() string {
	wat := make([]string, 0)
	for _, t := range i.indicator.Thresholds {
		var comparator string
		switch {
		case t.Operator <= indicator.LessThanOrEqualTo:
			comparator = "lt"
		case t.Operator >= indicator.GreaterThanOrEqualTo:
			comparator = "gt"
		default:
			log.Printf(
				"grafana dashboards only support lt/gt thresholds, threshold skipped: %s: %s %s %v\n",
				i.indicator.Name,
				t.Level,
				t.GetComparatorAbbrev(),
				t.Value,
			)
			continue
		}

		var level string
		switch t.Level {
		case "warning":
			level = t.Level
		case "critical":
			level = t.Level
		default:
			level = "custom"
		}
		wat = append(wat, fmt.Sprintf(`{"value":%v, "colorMode":"%v", "op":"%v", "fill":true, "line":true, "yaxis":"left"}`, t.Value, level, comparator))
	}
	return fmt.Sprintf(`[%s]`, strings.Join(wat, ","))
}
