package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

func TestMetadataOverride(t *testing.T) {
	t.Run("it overrides matching keys", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"deployment": "<%= spec.deployment %>",
					"source_id":  "<%= spec.job.name %>",
				},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
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

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"deployment": "<%= spec.deployment %>",
					"source_id":  "<%= spec.job.name %>",
				},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
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

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: []v1alpha1.IndicatorSpec{
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

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: []v1alpha1.IndicatorSpec{
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

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"ste": "foo"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "indicator-protocol",
					Version: "1.0",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "quest_rate",
					PromQL: "rate[$step]",
				}},
			},
		}

		doc.Interpolate()

		g.Expect(doc.Spec.Indicators[0].PromQL).To(Equal("rate[$step]"))
	})
}
