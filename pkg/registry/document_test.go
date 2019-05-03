package registry_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestDocumentTranslation(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		g := NewGomegaWithT(t)

		apiV0Doc := registry.APIV0Document{
			APIVersion: "apiV3",
			Product: registry.APIV0Product{
				Name:    "important-application",
				Version: "1.0",
			},
			Metadata: map[string]string{
				"someKey": "someValue",
			},
			Indicators: []registry.APIV0Indicator{{
				Name:   "performance-indicator",
				PromQL: "someQuery",
				Thresholds: []registry.APIV0Threshold{{
					Level:    "warning",
					Operator: "lte",
					Value:    100,
				}},
				Alert: registry.APIV0Alert{
					For:  "30s",
					Step: "5s",
				},
				Documentation: map[string]string{
					"anotherKey": "anotherValue",
				},
				Presentation: registry.APIV0Presentation{
					ChartType:    "bar",
					CurrentValue: false,
					Frequency:    50,
					Labels:       []string{"radical"},
				},
			}},
			Layout: registry.APIV0Layout{
				Title:       "The Important App",
				Description: "???",
				Sections: []registry.APIV0Section{
					{
						Title:       "The performance indicator",
						Description: "Pay attention!",
						Indicators:  []string{"performance-indicator"},
					},
				},
				Owner: "Waldo",
			},
		}

		indicatorDoc := registry.ToIndicatorDocument(apiV0Doc)

		g.Expect(indicatorDoc).To(Equal(indicator.Document{
			APIVersion: "apiV3",
			Product: indicator.Product{
				Name:    "important-application",
				Version: "1.0",
			},
			Metadata: map[string]string{
				"someKey": "someValue",
			},
			Indicators: []indicator.Indicator{{
				Name:   "performance-indicator",
				PromQL: "someQuery",
				Thresholds: []indicator.Threshold{{
					Level:    "warning",
					Operator: indicator.LessThanOrEqualTo,
					Value:    100,
				}},
				Alert: indicator.Alert{
					For:  "30s",
					Step: "5s",
				},
				Documentation: map[string]string{
					"anotherKey": "anotherValue",
				},
				Presentation: indicator.Presentation{
					ChartType:    "bar",
					CurrentValue: false,
					Frequency:    50,
					Labels:       []string{"radical"},
				},
			}},
			Layout: indicator.Layout{
				Title:       "The Important App",
				Description: "???",
				Sections: []indicator.Section{
					{
						Title:       "The performance indicator",
						Description: "Pay attention!",
						Indicators: []string{"performance-indicator"},
					},
				},
				Owner: "Waldo",
			},
		}))

	})
}
