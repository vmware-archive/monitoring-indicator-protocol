package v1_test

import (
	"testing"

	. "github.com/onsi/gomega"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestPopulateDefaults(t *testing.T) {
	t.Run("Default alert", func(t *testing.T) {
		t.Run("populates default alert config when no alert given", func(t *testing.T) {
			g := NewGomegaWithT(t)
			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Product:      "product",
							Name:         "well-performing-component",
							Type:         v1.DefaultIndicator,
							PromQL:       "promql()",
							Presentation: test_fixtures.DefaultPresentation(),
							Thresholds: []v1.Threshold{
								{
									Level:    "warning",
									Operator: v1.LessThan,
									Value:    10,
								},
							},
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)

			g.Expect(doc.Spec.Indicators[0].Thresholds[0].Alert).To(Equal(v1.Alert{
				For:  "1m",
				Step: "1m",
			}))
		})

		t.Run("populates default alert 'for' k/v when no alert for given", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "test_indicator",
							PromQL: "promql_query",
							Thresholds: []v1.Threshold{
								{
									Level:    "warning",
									Operator: v1.LessThan,
									Value:    10,
									Alert: v1.Alert{
										Step: "5m",
									},
								},
							},
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)

			g.Expect(doc.Spec.Indicators[0].Thresholds[0].Alert).To(Equal(v1.Alert{
				For:  "1m",
				Step: "5m",
			}))
		})

		t.Run("populates default alert 'step' k/v when no alert step given", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "test_indicator",
							PromQL: "promql_query",
							Thresholds: []v1.Threshold{
								{
									Level:    "warning",
									Operator: v1.LessThan,
									Value:    10,
									Alert: v1.Alert{
										For: "5m",
									},
								},
							},
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)

			g.Expect(doc.Spec.Indicators[0].Thresholds[0].Alert).To(Equal(v1.Alert{
				For:  "5m",
				Step: "1m",
			}))
		})

		t.Run("populates default alert for multiple thresholds", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "test_indicator",
							PromQL: "promql_query",
							Thresholds: []v1.Threshold{
								{
									Level:    "warning",
									Operator: v1.LessThan,
									Value:    10,
									Alert: v1.Alert{
										For: "5m",
									},
								},
								{
									Level:    "warning",
									Operator: v1.LessThan,
									Value:    10,
									Alert: v1.Alert{
										Step: "5m",
									},
								},
							},
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)

			g.Expect(doc.Spec.Indicators[0].Thresholds[0].Alert).To(Equal(v1.Alert{
				For:  "5m",
				Step: "1m",
			}))
			g.Expect(doc.Spec.Indicators[0].Thresholds[1].Alert).To(Equal(v1.Alert{
				For:  "1m",
				Step: "5m",
			}))
		})
	})

	t.Run("Default Presentation", func(t *testing.T) {
		t.Run("Sets a default presentation when non provided", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "well-performing-indicator-1",
							PromQL: "promql()",
						},
						{
							Name:   "well-performing-indicator-2",
							PromQL: "promql()",
							Presentation: v1.Presentation{
								CurrentValue: true,
							},
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)

			g.Expect(doc.Spec.Indicators[0].Presentation).To(BeEquivalentTo(v1.Presentation{
				ChartType:    v1.StepChart,
				CurrentValue: false,
				Frequency:    0,
				Labels:       []string{},
			}))
			g.Expect(doc.Spec.Indicators[1].Presentation).To(BeEquivalentTo(v1.Presentation{
				ChartType:    v1.StepChart,
				CurrentValue: true,
				Frequency:    0,
				Labels:       []string{},
			}))
		})
	})

	t.Run("Default layout", func(t *testing.T) {
		t.Run("Sets a default layout when non provided", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Product: v1.Product{
						Name:    "amazing-product",
						Version: "0.0.1",
					},
					Indicators: []v1.IndicatorSpec{
						{
							Product:      "product",
							Name:         "well-performing-indicator",
							Type:         v1.DefaultIndicator,
							PromQL:       "promql()",
							Presentation: test_fixtures.DefaultPresentation(),
							Thresholds: []v1.Threshold{
								{
									Level:    "warning",
									Operator: v1.LessThan,
									Value:    10,
								},
							},
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)

			g.Expect(doc.Spec.Layout).To(Equal(v1.Layout{
				Title: "amazing-product - 0.0.1",
				Sections: []v1.Section{{
					Title: "Metrics",
					Indicators: []string{
						"well-performing-indicator",
					},
				}},
			}))
		})

		t.Run("Does not fill the layout if a layout is provided", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Product:      "product",
							Name:         "name",
							Type:         v1.DefaultIndicator,
							PromQL:       "promql()",
							Presentation: test_fixtures.DefaultPresentation(),
						},
					},
					Layout: v1.Layout{
						Owner:       "foo",
						Title:       "asdf",
						Description: "description",
						Sections: []v1.Section{{
							Title:       "Section1",
							Description: "A section!",
							Indicators:  []string{"name"},
						}},
					},
				},
			}
			before := *doc.DeepCopy()
			v1.PopulateDefaults(&doc)
			g.Expect(before).To(Equal(doc))
		})

		t.Run("Fills the layout with SLI, then KPI, then other", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name: "name1",
							Type: v1.DefaultIndicator,
						}, {
							Name: "name2",
							Type: v1.ServiceLevelIndicator,
						}, {
							Name: "name3",
							Type: v1.KeyPerformanceIndicator,
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)
			layout := doc.Spec.Layout

			g.Expect(len(layout.Sections)).To(Equal(3))
			g.Expect(layout.Sections[0].Title).To(Equal("Service Level Indicators"))
			g.Expect(layout.Sections[1].Title).To(Equal("Key Performance Indicators"))
			g.Expect(layout.Sections[2].Title).To(Equal("Metrics"))
		})

		t.Run("If only two types of indicator, makes two sections", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name: "name1",
							Type: v1.DefaultIndicator,
						}, {
							Name: "name2",
							Type: v1.ServiceLevelIndicator,
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)
			layout := doc.Spec.Layout

			g.Expect(len(layout.Sections)).To(Equal(2))
			g.Expect(layout.Sections[0].Title).To(Equal("Service Level Indicators"))
			g.Expect(layout.Sections[1].Title).To(Equal("Metrics"))
		})

		t.Run("If only given one type of indicator, only makes one section", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name: "name",
							Type: v1.ServiceLevelIndicator,
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)
			layout := doc.Spec.Layout

			g.Expect(len(layout.Sections)).To(Equal(1))
			g.Expect(layout.Sections[0].Title).To(Equal("Service Level Indicators"))
		})

		t.Run("If only given other-type indicators, the section should be titled \"Indicators\"", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Indicators: []v1.IndicatorSpec{
						{
							Name: "name2",
							Type: v1.DefaultIndicator,
						},
					},
				},
			}

			v1.PopulateDefaults(&doc)
			layout := doc.Spec.Layout

			g.Expect(len(layout.Sections)).To(Equal(1))
			g.Expect(layout.Sections[0].Title).To(Equal("Metrics"))
		})

		t.Run("If given a layout, but no title, fills the title", func(t *testing.T) {
			g := NewGomegaWithT(t)

			doc := v1.IndicatorDocument{
				Spec: v1.IndicatorDocumentSpec{
					Product: v1.Product{
						Name:    "product",
						Version: "v1.2.4",
					},
					Indicators: []v1.IndicatorSpec{
						{
							Name:   "name",
							Type:   v1.DefaultIndicator,
							PromQL: "promql()",
						},
					},
					Layout: v1.Layout{
						Owner:       "foo",
						Title:       "",
						Description: "description",
						Sections: []v1.Section{{
							Title:       "Section1",
							Description: "A section!",
							Indicators:  []string{"name"},
						}},
					},
				},
			}
			v1.PopulateDefaults(&doc)
			g.Expect(doc.Spec.Layout.Title).To(Equal("product - v1.2.4"))
		})
	})

	t.Run("If fully complete, not changed", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "product",
					Version: "v1.2.4",
				},
				Indicators: []v1.IndicatorSpec{
					{
						Name:         "name",
						Type:         v1.DefaultIndicator,
						PromQL:       "promql()",
						Presentation: test_fixtures.DefaultPresentation(),
					},
				},
				Layout: v1.Layout{
					Owner:       "foo",
					Title:       "title!",
					Description: "description",
					Sections: []v1.Section{{
						Title:       "Section1",
						Description: "A section!",
						Indicators:  []string{"name"},
					}},
				},
			},
		}
		before := *doc.DeepCopy()
		v1.PopulateDefaults(&doc)
		g.Expect(before).To(Equal(doc))

	})
}
