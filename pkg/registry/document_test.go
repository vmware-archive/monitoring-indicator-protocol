package registry

import (
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	. "github.com/onsi/gomega"
	"testing"
)

func TestToAPIV0Document(t *testing.T) {
	t.Run("document has sections", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{},
			Metadata:   nil,
			Indicators: []indicator.Indicator{{
				Name:          "test_indicator",
				PromQL:        "test_indicator_promql{}",
			}},
			Layout: indicator.Layout{
				Title:       "test title",
				Sections: []indicator.Section{{
					Title:       "test section",
					Description: "test section description",
					Indicators:  []indicator.Indicator{{
						Name:          "test_indicator",
						PromQL:        "test_indicator_promql{}",
					}},
				}},
			},
		}

		result := ToAPIV0Document(doc)

		g.Expect(result.Layout.Sections).To(HaveLen(1))
		g.Expect(result.Layout.Sections[0].Title).To(Equal("test section"))
		g.Expect(result.Layout.Sections[0].Indicators).To(ConsistOf("test_indicator"))
	})
}
