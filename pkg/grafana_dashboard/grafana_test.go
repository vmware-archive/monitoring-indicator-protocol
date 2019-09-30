package grafana_dashboard_test

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
)

func TestDocumentToDashboard(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)
		document := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{
					{
						Name:          "test_indicator",
						PromQL:        `sum_over_time(gorouter_latency_ms[30m])`,
						Documentation: map[string]string{"title": "Test Indicator Title"},
						Thresholds: []v1.Threshold{{
							Level:    "critical",
							Operator: v1.GreaterThan,
							Value:    1000,
						}, {
							Level:    "warning",
							Operator: v1.LessThanOrEqualTo,
							Value:    700,
						}},
					},
					{
						Name:       "second_test_indicator",
						PromQL:     `rate(gorouter_requests[1m])`,
						Thresholds: []v1.Threshold{},
					},
				},
				Layout: v1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.UndefinedType)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(*dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
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

	t.Run("uses the layout information to create distinct rows", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{
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
				Layout: v1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1.Section{
						{
							Title:      "foo",
							Indicators: []string{"second_test_indicator"},
						},
						{
							Title:      "bar",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.UndefinedType)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(dashboard.Rows[0].Title).To(Equal("foo"))
		g.Expect(dashboard.Rows[0].Panels[0].Title).To(Equal("second_test_indicator"))
		g.Expect(dashboard.Rows[1].Title).To(Equal("bar"))
		g.Expect(dashboard.Rows[1].Panels[0].Title).To(Equal("Test Indicator Title"))
	})

	t.Run("replaces $step with $__interval", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		indicators := []v1.IndicatorSpec{
			{
				Name:   "test_indicator",
				PromQL: `rate(sum_over_time(gorouter_latency_ms[$step])[$step])`,
			},
		}
		document := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: indicators,
				Layout: v1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.UndefinedType)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(dashboard.Rows[0].Panels[0].Targets[0].Expression).To(BeEquivalentTo(`rate(sum_over_time(gorouter_latency_ms[$__interval])[$__interval])`))
	})

	t.Run("replaces only $step with $__interval", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		indicators := []v1.IndicatorSpec{
			{
				Name:   "test_indicator",
				PromQL: `rate(sum_over_time(gorouter_latency_ms[$steper])[$STEP])`,
			},
			{
				Name:   "another_indicator",
				PromQL: `avg_over_time(demo_latency{source_id="$step"}[5m])`,
			},
		}
		document := v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"ste": "123"},
			},
			Spec: v1.IndicatorDocumentSpec{

				Indicators: indicators,
				Layout: v1.Layout{
					Title: "Indicator Test Dashboard",
					Sections: []v1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator", "another_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.UndefinedType)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(dashboard.Rows[0].Panels[0].Targets[0].Expression).To(BeEquivalentTo(`rate(sum_over_time(gorouter_latency_ms[$steper])[$__interval])`))
		g.Expect(dashboard.Rows[0].Panels[1].Targets[0].Expression).ToNot(BeEquivalentTo(`avg_over_time(demo_latency{source_id="123p"}[5m])`))
	})

	t.Run("creates a filename based on product name and contents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Alert:  test_fixtures.DefaultAlert(),
					Thresholds: []v1.Threshold{{
						Level:    "critical",
						Operator: v1.LessThan,
						Value:    5,
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Test Indicator Title"},
				}},
				Layout: v1.Layout{
					Title: "Test Dashboard",
					Sections: []v1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		docBytes, err := json.Marshal(document)
		g.Expect(err).ToNot(HaveOccurred())
		filename := grafana_dashboard.DashboardFilename(docBytes, "test_product")
		// Should have a SHA1 in the middle, but don't want to specify the SHA
		g.Expect(filename).To(MatchRegexp("test_product_[a-f0-9]{40}\\.json"))
	})

	t.Run("includes annotations based on product & metadata alerts", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Alert:  test_fixtures.DefaultAlert(),
					Thresholds: []v1.Threshold{{
						Level:    "critical",
						Operator: v1.LessThan,
						Value:    5,
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Test Indicator Title"},
				}, {
					Name:   "second_test_indicator",
					PromQL: "second_test_query",
					Alert:  test_fixtures.DefaultAlert(),
					Thresholds: []v1.Threshold{{
						Level:    "critical",
						Operator: v1.GreaterThan,
						Value:    10,
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Second Test Indicator Title"},
				}},
				Layout: v1.Layout{
					Title: "Test Dashboard",
					Sections: []v1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator", "second_test_indicator"},
						},
					},
				},
			},
		}

		dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.UndefinedType)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(dashboard.Annotations.List).To(ConsistOf(grafana_dashboard.GrafanaAnnotation{
			Enable:      true,
			Expr:        "ALERTS{product=\"test_product\",deployment=\"test_deployment\"}",
			TagKeys:     "level",
			TitleFormat: "{{alertname}} is {{alertstate}} in the {{level}} threshold",
			IconColor:   "#1f78c1",
		}))
	})

	t.Run("generates dashboards for indicator type requested", func(t *testing.T) {
		t.Run("sli", func(t *testing.T) {
			buffer := bytes.NewBuffer(nil)
			log.SetOutput(buffer)

			g := NewGomegaWithT(t)
			document := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "sli_test",
							PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
							Type:   v1.ServiceLevelIndicator,
						},
						{
							Name:   "kpi_test",
							Type:   v1.KeyPerformanceIndicator,
							PromQL: `key_test`,
						},
						{
							Name:   "another_sli_test",
							Type:   v1.ServiceLevelIndicator,
							PromQL: `super_test`,
						},
					},
					Layout: v1.Layout{
						Title: "Indicator Test Dashboard",
						Sections: []v1.Section{
							{
								Title:      "Test Section Title",
								Indicators: []string{"sli_test", "kpi_test", "another_sli_test"},
							},
						},
					},
				},
			}

			dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.ServiceLevelIndicator)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(*dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
				Title: "Indicator Test Dashboard",
				Rows: []grafana_dashboard.GrafanaRow{{
					Title: "Test Section Title",
					Panels: []grafana_dashboard.GrafanaPanel{
						{
							Title: "sli_test",
							Type:  "graph",
							Targets: []grafana_dashboard.GrafanaTarget{{
								Expression: `sum_over_time(gorouter_latency_ms[30m])`,
							}},
						},
						{
							Title: "another_sli_test",
							Type:  "graph",
							Targets: []grafana_dashboard.GrafanaTarget{{
								Expression: `super_test`,
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

		t.Run("kpi", func(t *testing.T) {
			buffer := bytes.NewBuffer(nil)
			log.SetOutput(buffer)

			g := NewGomegaWithT(t)
			document := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "sli_test",
							PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
							Type:   v1.ServiceLevelIndicator,
						},
						{
							Name:   "kpi_test",
							Type:   v1.KeyPerformanceIndicator,
							PromQL: `key_test`,
						},
						{
							Name:   "another_sli_test",
							PromQL: `wow`,
							Type:   v1.ServiceLevelIndicator,
						},
					},
					Layout: v1.Layout{
						Title: "Indicator Test Dashboard",
						Sections: []v1.Section{
							{
								Title:      "Test Section Title",
								Indicators: []string{"sli_test", "kpi_test", "another_sli_test"},
							},
						},
					},
				},
			}

			dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.KeyPerformanceIndicator)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(*dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
				Title: "Indicator Test Dashboard",
				Rows: []grafana_dashboard.GrafanaRow{{
					Title: "Test Section Title",
					Panels: []grafana_dashboard.GrafanaPanel{
						{
							Title: "kpi_test",
							Type:  "graph",
							Targets: []grafana_dashboard.GrafanaTarget{{
								Expression: `key_test`,
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

		t.Run("other", func(t *testing.T) {
			buffer := bytes.NewBuffer(nil)
			log.SetOutput(buffer)

			g := NewGomegaWithT(t)
			document := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "sli_test",
							PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
							Type:   v1.ServiceLevelIndicator,
						},
						{
							Name:   "other",
							PromQL: `other_test`,
							Type:   v1.DefaultIndicator,
						},
					},
					Layout: v1.Layout{
						Title: "Indicator Test Dashboard",
						Sections: []v1.Section{
							{
								Title:      "Test Section Title",
								Indicators: []string{"sli_test", "other"},
							},
						},
					},
				},
			}

			dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.DefaultIndicator)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(*dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
				Title: "Indicator Test Dashboard",
				Rows: []grafana_dashboard.GrafanaRow{{
					Title: "Test Section Title",
					Panels: []grafana_dashboard.GrafanaPanel{
						{
							Title: "other",
							Type:  "graph",
							Targets: []grafana_dashboard.GrafanaTarget{{
								Expression: `other_test`,
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

		t.Run("all", func(t *testing.T) {
			buffer := bytes.NewBuffer(nil)
			log.SetOutput(buffer)

			g := NewGomegaWithT(t)
			document := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "sli_test",
							PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
							Type:   v1.ServiceLevelIndicator,
						},
						{
							Name:   "kpi_test",
							Type:   v1.KeyPerformanceIndicator,
							PromQL: `key_test`,
						},
						{
							Name:   "other",
							PromQL: `wow`,
							Type:   v1.DefaultIndicator,
						},
					},
					Layout: v1.Layout{
						Title: "Indicator Test Dashboard",
						Sections: []v1.Section{
							{
								Title:      "Test Section Title",
								Indicators: []string{"sli_test", "kpi_test", "other"},
							},
						},
					},
				},
			}

			dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.UndefinedType)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(*dashboard).To(BeEquivalentTo(grafana_dashboard.GrafanaDashboard{
				Title: "Indicator Test Dashboard",
				Rows: []grafana_dashboard.GrafanaRow{{
					Title: "Test Section Title",
					Panels: []grafana_dashboard.GrafanaPanel{
						{
							Title: "sli_test",
							Type:  "graph",
							Targets: []grafana_dashboard.GrafanaTarget{{
								Expression: `sum_over_time(gorouter_latency_ms[30m])`,
							}},
						},
						{
							Title: "kpi_test",
							Type:  "graph",
							Targets: []grafana_dashboard.GrafanaTarget{{
								Expression: `key_test`,
							}},
						},
						{
							Title: "other",
							Type:  "graph",
							Targets: []grafana_dashboard.GrafanaTarget{{
								Expression: `wow`,
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

		t.Run("return nil when no indicators match the requested type", func(t *testing.T) {
			buffer := bytes.NewBuffer(nil)
			log.SetOutput(buffer)

			g := NewGomegaWithT(t)
			document := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "kpi_test",
							Type:   v1.KeyPerformanceIndicator,
							PromQL: `key_test`,
						},
						{
							Name:   "other",
							PromQL: `wow`,
							Type:   v1.DefaultIndicator,
						},
					},
					Layout: v1.Layout{
						Title: "Indicator Test Dashboard",
						Sections: []v1.Section{
							{
								Title:      "Test Section Title",
								Indicators: []string{"kpi_test", "other"},
							},
						},
					},
				},
			}

			dashboard, err := grafana_dashboard.DocumentToDashboard(document, v1.ServiceLevelIndicator)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(dashboard).To(BeNil())
		})
	})
}
