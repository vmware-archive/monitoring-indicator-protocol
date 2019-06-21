package indicator_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func TestMetadataOverride(t *testing.T) {
	t.Run("it overrides matching keys", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "indicator-protocol",
				Version: "1.0",
			},
			Metadata: map[string]string{
				"deployment": "<%= spec.deployment %>",
				"source_id":  "<%= spec.job.name %>",
			},
		}
		doc.OverrideMetadata(map[string]string{
			"source_id":  "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		})

		g.Expect(doc.Metadata).To(BeEquivalentTo(map[string]string{
			"source_id":  "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		}))
	})

	t.Run("it adds non-matching keys", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "indicator-protocol",
				Version: "1.0",
			},
			Metadata: map[string]string{
				"deployment": "<%= spec.deployment %>",
				"source_id":  "<%= spec.job.name %>",
			},
		}
		doc.OverrideMetadata(map[string]string{
			"force_id":   "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		})

		g.Expect(doc.Metadata).To(BeEquivalentTo(map[string]string{
			"source_id":  "<%= spec.job.name %>",
			"force_id":   "indicator-protocol-lol",
			"deployment": "indicator-protocol-wow",
		}))
	})
}

func TestMetadataInterpolation(t *testing.T) {
	t.Run("it replaces $metadata in promql with the metadata value", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := indicator.Document{
			Metadata: map[string]string{
				"foo": "bar",
			},
			Indicators: []indicator.Indicator{
				{
					PromQL: "something $foo something",
				},
			},
		}

		doc.Interpolate()

		g.Expect(doc.Indicators[0].PromQL).To(Equal("something bar something"))
	})

	t.Run("it doesn't replace $metadata in other fields", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := indicator.Document{
			Metadata: map[string]string{
				"foo": "bar",
			},
			Indicators: []indicator.Indicator{
				{
					Documentation: map[string]string{"baz": "something $foo something"},
				},
			},
		}

		doc.Interpolate()

		g.Expect(doc.Indicators[0].Documentation["baz"]).To(Equal("something $foo something"))
	})

	t.Run("it does not replace $step, even if a partial key is present", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "indicator-protocol",
				Version: "1.0",
			},
			Metadata: map[string]string{"ste": "foo"},
			Indicators: []indicator.Indicator{{
				Name:   "quest_rate",
				PromQL: "rate[$step]",
			}},
		}

		doc.Interpolate()

		g.Expect(doc.Indicators[0].PromQL).To(Equal("rate[$step]"))
	})
}
