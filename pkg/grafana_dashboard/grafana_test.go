package grafana_dashboard_test

import (
	"bytes"
	"log"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
)

func TestDocumentToDashboard(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	g := NewGomegaWithT(t)

	document := indicator.Document{
		Indicators: []indicator.Indicator{
			{
				Name:          "test_indicator",
				PromQL:        `sum_over_time(gorouter_latency_ms[30m])`,
				Documentation: map[string]string{"title": "Test Indicator Title"},
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.GreaterThan,
					Value:    1000,
				}, {
					Level:    "warning",
					Operator: indicator.LessThanOrEqualTo,
					Value:    700,
				}},
			},
			{
				Name:       "second_test_indicator",
				PromQL:     `rate(gorouter_requests[1m])`,
				Thresholds: []indicator.Threshold{},
			},
		},
		Layout: indicator.Layout{
			Title: "Indicator Test Dashboard",
		},
	}

	dashboard := grafana_dashboard.DocumentToDashboard(document)

	g.Expect(dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
		Title: "Indicator Test Dashboard",
		Rows: []grafana_dashboard.GrafanaRow{{
			Title: "Test Indicator Title",
			Panels: []grafana_dashboard.GrafanaPanel{{
				Title: "Test Indicator Title",
				Type:  "graph",
				Targets: []grafana_dashboard.GrafanaTarget{{
					Expression: `sum_over_time(gorouter_latency_ms[30m])`,
				}},
				Thresholds: []grafana_dashboard.GrafanaThreshold{{
					Value:     1000,
					ColorMode: "critical",
					Op:        "gt",
					Fill:      true,
					Line:      true,
					Yaxis:     "left",
				}, {
					Value:     700,
					ColorMode: "warning",
					Op:        "lt",
					Fill:      true,
					Line:      true,
					Yaxis:     "left",
				}},
			}},
		}, {
			Title: "second_test_indicator",
			Panels: []grafana_dashboard.GrafanaPanel{{
				Title: "second_test_indicator",
				Type:  "graph",
				Targets: []grafana_dashboard.GrafanaTarget{{
					Expression: `rate(gorouter_requests[1m])`,
				}},
			}},
		}},
	}))
}
