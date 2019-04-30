package indicator_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func TestUpdateMetadata(t *testing.T) {
	fileBytes := []byte(`---
apiVersion: v0
product: 
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: well-performing-deployment

indicators:
- name: test_performance_indicator
  promql: query_metric{source_id="$deployment"}
`)

	t.Run("it replaces promql $EXPR with metadata tags", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument(fileBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(BeEquivalentTo(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Metadata:   map[string]string{"deployment": "well-performing-deployment"},
			Indicators: []indicator.Indicator{
				{
					Name:   "test_performance_indicator",
					PromQL: `query_metric{source_id="well-performing-deployment"}`,
					Alert: indicator.Alert{
						For:  "1m",
						Step: "1m",
					},
					Presentation: indicator.Presentation{
						CurrentValue: false,
						ChartType:    "step",
						Frequency:    0,
						Labels:       []string{},
					},
				},
			},
			Layout: indicator.Layout{
				Sections: []indicator.Section{{
					Title: "Metrics",
					Indicators: []indicator.Indicator{
						{
							Name:   "test_performance_indicator",
							PromQL: `query_metric{source_id="well-performing-deployment"}`,
							Alert: indicator.Alert{
								For:  "1m",
								Step: "1m",
							},
							Presentation: indicator.Presentation{
								CurrentValue: false,
								ChartType:    "step",
								Frequency:    0,
								Labels:       []string{},
							},
						},
					},
				}},
			},
		}))
	})

	t.Run("it does not replaces promql $EXPR with metadata tags when passed flag", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument(fileBytes, indicator.SkipMetadataInterpolation)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(Equal(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Metadata:   map[string]string{"deployment": "well-performing-deployment"},
			Indicators: []indicator.Indicator{
				{
					Name:   "test_performance_indicator",
					PromQL: `query_metric{source_id="$deployment"}`,
					Alert: indicator.Alert{
						For:  "1m",
						Step: "1m",
					},
					Presentation: indicator.Presentation{
						CurrentValue: false,
						ChartType:    "step",
						Frequency:    0,
						Labels:       []string{},
					},
				},
			},
			Layout: indicator.Layout{
				Sections: []indicator.Section{{
					Title: "Metrics",
					Indicators: []indicator.Indicator{
						{
							Name:   "test_performance_indicator",
							PromQL: `query_metric{source_id="$deployment"}`,
							Alert: indicator.Alert{
								For:  "1m",
								Step: "1m",
							},
							Presentation: indicator.Presentation{
								CurrentValue: false,
								ChartType:    "step",
								Frequency:    0,
								Labels:       []string{},
							},
						},
					},
				}},
			},
		}))
	})
}
