package v1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestMetadataOverride(t *testing.T) {
	t.Run("it overrides matching keys", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"deployment": "<%= spec.deployment %>",
					"source_id":  "<%= spec.job.name %>",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "indicator-protocol",
					Version: "1.0",
				},
			},
		}
		doc.OverrideMetadata(map[string]string{
			"source_id":  "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		})

		g.Expect(doc.ObjectMeta.Labels).To(BeEquivalentTo(map[string]string{
			"source_id":  "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		}))
	})

	t.Run("it adds non-matching keys", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"deployment": "<%= spec.deployment %>",
					"source_id":  "<%= spec.job.name %>",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "indicator-protocol",
					Version: "1.0",
				},
			},
		}
		doc.OverrideMetadata(map[string]string{
			"force_id":   "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		})

		g.Expect(doc.ObjectMeta.Labels).To(BeEquivalentTo(map[string]string{
			"source_id":  "<%= spec.job.name %>",
			"force_id":   "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		}))
	})
}

func TestMetadataInterpolation(t *testing.T) {
	t.Run("it replaces $metadata in promql with the metadata value", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{
					{
						PromQL: "something $foo something",
					},
				},
			},
		}

		doc.Interpolate()

		g.Expect(doc.Spec.Indicators[0].PromQL).To(Equal("something bar something"))
	})

	t.Run("it doesn't replace $metadata in other fields", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{
					{
						Documentation: map[string]string{"baz": "something $foo something"},
					},
				},
			},
		}

		doc.Interpolate()

		g.Expect(doc.Spec.Indicators[0].Documentation["baz"]).To(Equal("something $foo something"))
	})

	t.Run("it does not replace $step, even if a partial key is present", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"ste": "foo"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "indicator-protocol",
					Version: "1.0",
				},
				Indicators: []v1.IndicatorSpec{{
					Name:   "quest_rate",
					PromQL: "rate[$step]",
				}},
			},
		}

		doc.Interpolate()

		g.Expect(doc.Spec.Indicators[0].PromQL).To(Equal("rate[$step]"))
	})
}

func TestFillLayout(t *testing.T) {
	t.Run("Does not fill the layout if a layout is provided", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{
					{
						Product: "product",
						Name:    "name",
						Type:    v1.DefaultIndicator,
						PromQL:  "promql()",
						Alert:         test_fixtures.DefaultAlert(),
						Presentation:  test_fixtures.DefaultPresentation(),
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
						Name:          "name",
						Type:          v1.DefaultIndicator,
						PromQL:        "promql()",
						Alert:         test_fixtures.DefaultAlert(),
						Presentation:  test_fixtures.DefaultPresentation(),
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
