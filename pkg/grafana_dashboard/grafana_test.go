package grafana_dashboard_test

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func TestDocumentToDashboard(t *testing.T) {
	t.Run("works", func(t *testing.T) {
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
				Sections: []indicator.Section{
					{
						Title: "Test Section Title",
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
					},
				},
			},
		}

		dashboard := grafana_dashboard.DocumentToDashboard(document)

		g.Expect(dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
			Title: "Indicator Test Dashboard",
			Rows: []grafana_dashboard.GrafanaRow{{
				Title: "Test Section Title",
				Panels: []grafana_dashboard.GrafanaPanel{
					{
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
					},
					{
						Title: "second_test_indicator",
						Type:  "graph",
						Targets: []grafana_dashboard.GrafanaTarget{{
							Expression: `rate(gorouter_requests[1m])`,
						}},
					},
				},
			}},
			Annotations: grafana_dashboard.GrafanaAnnotations{
				List: []grafana_dashboard.GrafanaAnnotation{
					{
						Enable:      true,
						Expr:        "ALERTS{product=\"\"}",
						TagKeys:     "level",
						TitleFormat: "{{alertname}} is {{alertstate}} in the {{level}} threshold",
						IconColor:   "#1f78c1",
					},
				},
			},
		}))
	})

	t.Run("uses the IP layout information to create distinct rows", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		document := indicator.Document{
			Indicators: []indicator.Indicator{
				{
					Name:          "test_indicator",
					PromQL:        `sum_over_time(gorouter_latency_ms[30m])`,
					Documentation: map[string]string{"title": "Test Indicator Title"},
				},
				{
					Name:   "second_test_indicator",
					PromQL: `rate(gorouter_requests[1m])`,
				},
			},
			Layout: indicator.Layout{
				Title: "Indicator Test Dashboard",
				Sections: []indicator.Section{
					{
						Title: "foo",
						Indicators: []indicator.Indicator{
							{
								Name:   "second_test_indicator",
								PromQL: `rate(gorouter_requests[1m])`,
							},
						},
					},
					{
						Title: "bar",
						Indicators: []indicator.Indicator{
							{
								Name:          "test_indicator",
								PromQL:        `sum_over_time(gorouter_latency_ms[30m])`,
								Documentation: map[string]string{"title": "Test Indicator Title"},
							},
						},
					},
				},
			},
		}

		dashboard := grafana_dashboard.DocumentToDashboard(document)

		g.Expect(dashboard.Rows[0].Title).To(Equal("foo"))
		g.Expect(dashboard.Rows[0].Panels[0].Title).To(Equal("second_test_indicator"))
		g.Expect(dashboard.Rows[1].Title).To(Equal("bar"))
		g.Expect(dashboard.Rows[1].Panels[0].Title).To(Equal("Test Indicator Title"))
	})

	t.Run("falls back to product name/version when layout title is missing", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		document := indicator.Document{
			Product: indicator.Product{
				Name:    "test product",
				Version: "v0.9",
			},
			Layout: indicator.Layout{
				Sections: []indicator.Section{
					{
						Title: "test section",
						Indicators: []indicator.Indicator{
							{
								Name:   "test_indicator",
								PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
							},
						},
					},
				},
			},
		}

		dashboard := grafana_dashboard.DocumentToDashboard(document)

		g.Expect(dashboard.Title).To(BeEquivalentTo("test product - v0.9"))
	})

	t.Run("replaces $step with $__interval", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		document := indicator.Document{
			Indicators: []indicator.Indicator{
				{
					Name:   "test_indicator",
					PromQL: `sum_over_time(gorouter_latency_ms[$step])`,
				},
			},
			Layout: indicator.Layout{
				Title: "Indicator Test Dashboard",
				Sections: []indicator.Section{
					{
						Title: "Test Section Title",
						Indicators: []indicator.Indicator{
							{
								Name:   "test_indicator",
								PromQL: `rate(sum_over_time(gorouter_latency_ms[$step])[$step])`,
							},
						},
					},
				},
			},
		}

		dashboard := grafana_dashboard.DocumentToDashboard(document)

		g.Expect(dashboard.Rows[0].Panels[0].Targets[0].Expression).To(BeEquivalentTo(`rate(sum_over_time(gorouter_latency_ms[$__interval])[$__interval])`))
	})

	t.Run("creates a filename based on product name and contents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "test_product",
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: `test_query{deployment="test_deployment"}`,
				Alert:  test_fixtures.DefaultAlert(),
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.LessThan,
					Value:    5,
				}},
				Presentation:  test_fixtures.DefaultPresentation(),
				Documentation: map[string]string{"title": "Test Indicator Title"},
			}},
			Layout: indicator.Layout{
				Title: "Test Dashboard",
				Sections: []indicator.Section{
					{
						Title: "Test Section Title",
					},
				},
			},
		}
		document.Layout.Sections[0].Indicators = document.Indicators

		docBytes, err := json.Marshal(document)
		g.Expect(err).ToNot(HaveOccurred())
		filename := grafana_dashboard.DashboardFilename(docBytes, "test_product")
		g.Expect(filename).To(Equal("test_product_c47308287f60c776ba39f27aaa743bca7c6c6387.json"))
	})

	t.Run("includes annotations based on product & metadata alerts", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "test_product",
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: `test_query{deployment="test_deployment"}`,
				Alert:  test_fixtures.DefaultAlert(),
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.LessThan,
					Value:    5,
				}},
				Presentation:  test_fixtures.DefaultPresentation(),
				Documentation: map[string]string{"title": "Test Indicator Title"},
			}, {
				Name:   "second_test_indicator",
				PromQL: "second_test_query",
				Alert:  test_fixtures.DefaultAlert(),
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.GreaterThan,
					Value:    10,
				}},
				Presentation:  test_fixtures.DefaultPresentation(),
				Documentation: map[string]string{"title": "Second Test Indicator Title"},
			}},
			Layout: indicator.Layout{
				Title: "Test Dashboard",
				Sections: []indicator.Section{
					{
						Title: "Test Section Title",
					},
				},
			},
		}
		document.Layout.Sections[0].Indicators = document.Indicators

		dashboard := grafana_dashboard.DocumentToDashboard(document)

		g.Expect(dashboard.Annotations.List).To(ConsistOf(grafana_dashboard.GrafanaAnnotation{
			Enable:      true,
			Expr:        "ALERTS{product=\"test_product\",deployment=\"test_deployment\"}",
			TagKeys:     "level",
			TitleFormat: "{{alertname}} is {{alertstate}} in the {{level}} threshold",
			IconColor:   "#1f78c1",
		}))
	})
}
