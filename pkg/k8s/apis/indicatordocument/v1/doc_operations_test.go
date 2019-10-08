package v1_test

import (
	"testing"

	. "github.com/onsi/gomega"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestFillLayout(t *testing.T) {
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
