package registry

import (
	. "github.com/onsi/gomega"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"testing"
	"time"
)

func TestToAPIV0Document(t *testing.T) {
	t.Run("document has sections", func(t *testing.T) {
		g := NewGomegaWithT(t)

		doc := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{},
			Metadata:   nil,
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: "test_indicator_promql{}",
				Presentation: &indicator.Presentation{
					ChartType:    "step",
					CurrentValue: false,
					Frequency:    time.Minute,
				},
			}},
			Layout: indicator.Layout{
				Title: "test title",
				Sections: []indicator.Section{{
					Title:       "test section",
					Description: "test section description",
					Indicators: []indicator.Indicator{{
						Name:   "test_indicator",
						PromQL: "test_indicator_promql{}",
					}},
				}},
			},
		}

		result := ToAPIV0Document(doc)

		g.Expect(result.Layout.Sections).To(HaveLen(1))
		g.Expect(result.Layout.Sections[0].Title).To(Equal("test section"))
		g.Expect(result.Layout.Sections[0].Indicators).To(ConsistOf("test_indicator"))
		g.Expect(result.Indicators).To(ConsistOf(APIV0Indicator{
			Name:       "test_indicator",
			PromQL:     "test_indicator_promql{}",
			Thresholds: []APIV0Threshold{},
			Presentation: &APIV0Presentation{
				ChartType:    "step",
				CurrentValue: false,
				Frequency:    60,
			},
		}))
	})
}
