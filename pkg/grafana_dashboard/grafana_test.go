package grafana_dashboard_test

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	"github.com/grafana-tools/sdk"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestToGrafanaDashboard(t *testing.T) {

	t.Run("translate a single chart to the proper grafana json", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicator := v1.IndicatorSpec{
			Name:   "test_indicator",
			Type:   v1.UndefinedType,
			PromQL: `8`,
			Documentation: map[string]string{
				"title": "test title",
			},
			Presentation: v1.Presentation{
				ChartType: v1.UndefinedChart,
				Units:     "mbytes",
			},
		}

		data := grafana_dashboard.ToGrafanaPanel(indicator)

		panel := sdk.NewGraph("test_indicator")

		height := 10
		width := 24

		panel.GridPos = struct {
			H *int `json:"h,omitempty"`
			W *int `json:"w,omitempty"`
			X *int `json:"x,omitempty"`
			Y *int `json:"y,omitempty"`
		}{
			H: &height,
			W: &width,
		}
		panel.AddTarget(&sdk.Target{
			Expr: `8`,
		})
		panel.GraphPanel.Xaxis = sdk.Axis{
			Format: "time",
			Show:   true,
		}
		panel.GraphPanel.Yaxes = []sdk.Axis{
			{
				Format: "mbytes",
				Show:   true,
			},
			{
				Format: "mbytes",
			},
		}
		panel.GraphPanel.Lines = true
		panel.GraphPanel.Linewidth = 1
		panel.AliasColors = map[string]string{}

		stringPointer := `## title
test title

`
		panel.GraphPanel.Description = &stringPointer

		raw, _ := json.Marshal(data)
		println(string(raw))
		g.Expect(data).To(Equal(panel))
	})

	t.Run("translates several documents to grafana dashboard jsons", func(t *testing.T) {
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

		dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.UndefinedType)
		anotherDashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.UndefinedType)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(dashboard.ID).To(Equal(uint(0)))
		g.Expect(dashboard.Title).To(Equal("Indicator Test Dashboard"))
		g.Expect(dashboard.Panels).To(HaveLen(2))
		g.Expect(dashboard.Panels[1].Title).To(Equal("test_indicator"))
		g.Expect(dashboard.Panels[1].GraphPanel.Targets[0].Expr).To(Equal("sum_over_time(gorouter_latency_ms[30m])"))
		g.Expect(*dashboard.Panels[1].Description).To(ContainSubstring("Test Indicator Title"))
		g.Expect(dashboard.Panels[1].GraphPanel.Thresholds).To(HaveLen(2))

		g.Expect(anotherDashboard.ID).To(Equal(uint(0)))
		g.Expect(anotherDashboard.Title).To(Equal("Indicator Test Dashboard"))
		g.Expect(anotherDashboard.Panels).To(HaveLen(2))
		g.Expect(anotherDashboard.Panels[1].Title).To(Equal("test_indicator"))
		g.Expect(anotherDashboard.Panels[1].GraphPanel.Targets[0].Expr).To(Equal("sum_over_time(gorouter_latency_ms[30m])"))
		g.Expect(*anotherDashboard.Panels[1].Description).To(ContainSubstring("Test Indicator Title"))
		g.Expect(anotherDashboard.Panels[1].GraphPanel.Thresholds).To(HaveLen(2))
	})

	t.Run("takes indicator documentation and turns it into a graph description", func(t *testing.T) {
		g := NewGomegaWithT(t)

		description := `## description
This is a valid markdown description.

**Use**: This indicates nothing. It is placeholder text.

**Type**: Gauge
**Frequency**: 60 s

## measurement
Average latency over last 5 minutes per instance

## recommendedResponse
Panic! Run around in circles flailing your arms.

## thresholdNote
These are environment specific

## title
Doc Performance Indicator

`
		g.Expect(grafana_dashboard.ToGrafanaDescription(map[string]string{
			"title":       "Doc Performance Indicator",
			"measurement": "Average latency over last 5 minutes per instance",
			"description": `This is a valid markdown description.

**Use**: This indicates nothing. It is placeholder text.

**Type**: Gauge
**Frequency**: 60 s`,
			"recommendedResponse": "Panic! Run around in circles flailing your arms.",
			"thresholdNote":       "These are environment specific",
		})).To(Equal(&description))
	})

	t.Run("takes threshold and turns it into grafana threshold", func(t *testing.T) {
		g := NewGomegaWithT(t)

		thresholds := []v1.Threshold{
			{
				Level:    "critical",
				Operator: v1.LessThan,
				Value:    7,
			},
		}

		g.Expect(grafana_dashboard.ToGrafanaThresholds(thresholds)).To(Equal([]sdk.Threshold{{
			Value:     7,
			Op:        "lt",
			Line:      true,
			Fill:      true,
			ColorMode: "critical",
			Yaxis:     "left",
		}}))
	})

	t.Run("uses the layout information to generate rows", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{
					{
						Name:   "test_indicator",
						PromQL: `1`,
					},
					{
						Name:   "second_test_indicator",
						PromQL: `2`,
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

		dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.UndefinedType)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(dashboard.Panels).To(HaveLen(4))
		g.Expect(dashboard.Panels[0].Title).To(Equal("foo"))
		g.Expect(dashboard.Panels[1].Title).To(Equal("second_test_indicator"))
		g.Expect(dashboard.Panels[2].Title).To(Equal("bar"))
		g.Expect(dashboard.Panels[3].Title).To(Equal("test_indicator"))

		g.Expect(*dashboard.Panels[3].GridPos.Y).To(Equal(12))

		g.Expect(dashboard.Panels[0].ID).To(BeNumerically("==", 0))
		g.Expect(dashboard.Panels[3].ID).To(BeNumerically("==", 3))

		json.Marshal(dashboard)
	})

	t.Run(`replaces step with interval`, func(t *testing.T) {
		g := NewGomegaWithT(t)

		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

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

		dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.UndefinedType)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(dashboard.Panels[1].GraphPanel.Targets[0].Expr).To(BeEquivalentTo(`rate(sum_over_time(gorouter_latency_ms[$__interval])[$__interval])`))
	})

	t.Run("replaces only step with interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

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

		dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.UndefinedType)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(dashboard.Panels[1].GraphPanel.Targets[0].Expr).To(BeEquivalentTo(`rate(sum_over_time(gorouter_latency_ms[$steper])[$__interval])`))
		g.Expect(dashboard.Panels[2].GraphPanel.Targets[0].Expr).ToNot(BeEquivalentTo(`avg_over_time(demo_latency{source_id="123p"}[5m])`))
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
					Thresholds: []v1.Threshold{{
						Level:    "critical",
						Operator: v1.LessThan,
						Value:    5,
						Alert:    test_fixtures.DefaultAlert(),
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

			dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.ServiceLevelIndicator)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dashboard.Panels).To(HaveLen(3))
			g.Expect(dashboard.Panels[1].Title).To(Equal("sli_test"))
			g.Expect(dashboard.Panels[2].Title).To(Equal("another_sli_test"))
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

			dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.KeyPerformanceIndicator)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dashboard.Panels).To(HaveLen(2))
			g.Expect(dashboard.Panels[1].Title).To(Equal("kpi_test"))
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

			dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.DefaultIndicator)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(dashboard.Panels).To(HaveLen(2))
			g.Expect(dashboard.Panels[1].Title).To(Equal("other"))
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

			dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.UndefinedType)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dashboard.Panels).To(HaveLen(4))
			g.Expect(dashboard.Panels[1].Title).To(Equal("sli_test"))
			g.Expect(dashboard.Panels[2].Title).To(Equal("kpi_test"))
			g.Expect(dashboard.Panels[3].Title).To(Equal("other"))
		})

		t.Run("return nil when no layout provided", func(t *testing.T) {
			buffer := bytes.NewBuffer(nil)
			log.SetOutput(buffer)

			g := NewGomegaWithT(t)
			document := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "sli_test",
							PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
						},
						{
							Name:   "kpi_test",
							PromQL: `key_test`,
						},
						{
							Name:   "other",
							PromQL: `wow`,
						},
					},
				},
			}

			dashboard, err := grafana_dashboard.ToGrafanaDashboard(document, v1.ServiceLevelIndicator)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(dashboard).To(BeNil())
		})
	})
}
