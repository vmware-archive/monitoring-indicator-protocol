package registry_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestDocumentTranslation(t *testing.T) {
	t.Run("it translates", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicatorDoc := registry.APIDocumentResponse{
			APIVersion: api_versions.V1,
			Metadata: registry.APIMetadataResponse{
				Labels: map[string]string{
					"someKey": "someValue",
				},
			},
			Kind: "IndicatorDocument",
			Spec: registry.APIDocumentSpecResponse{
				Product: registry.APIProductResponse{
					Name:    "important-application",
					Version: "1.0",
				},
				Indicators: []registry.APIIndicatorResponse{{
					Name:   "performance-indicator",
					PromQL: "someQuery",
					Thresholds: []registry.APIThresholdResponse{{
						Level:    "warning",
						Operator: "lte",
						Value:    100,
					}},
					Alert: registry.APIAlertResponse{
						For:  "30s",
						Step: "5s",
					},
					Documentation: map[string]string{
						"anotherKey": "anotherValue",
					},
					Presentation: registry.APIPresentationResponse{
						ChartType:    "bar",
						CurrentValue: false,
						Frequency:    50,
						Labels:       []string{"radical"},
					},
					ServiceLevel: &registry.APIServiceLevelResponse{
						Objective: 100,
					},
				}},
				Layout: registry.APILayoutResponse{
					Title:       "The Important App",
					Description: "???",
					Sections: []registry.APISectionResponse{
						{
							Title:       "The performance indicator",
							Description: "Pay attention!",
							Indicators:  []string{"performance-indicator"},
						},
					},
					Owner: "Waldo",
				},
			},
		}

		g.Expect(registry.ToIndicatorDocument(indicatorDoc)).To(Equal(v1.IndicatorDocument{
			TypeMeta: metaV1.TypeMeta{
				Kind:       "IndicatorDocument",
				APIVersion: api_versions.V1,
			},
			ObjectMeta: metaV1.ObjectMeta{
				Labels: map[string]string{
					"someKey": "someValue",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "important-application",
					Version: "1.0",
				},
				Indicators: []v1.IndicatorSpec{{
					Name:   "performance-indicator",
					PromQL: "someQuery",
					Alert: v1.Alert{
						For:  "30s",
						Step: "5s",
					},
					Thresholds: []v1.Threshold{{
						Level:    "warning",
						Operator: v1.LessThanOrEqualTo,
						Value:    100,
					}},
					ServiceLevel: &v1.ServiceLevel{
						Objective: 100,
					},
					Documentation: map[string]string{
						"anotherKey": "anotherValue",
					},
					Presentation: v1.Presentation{
						ChartType:    "bar",
						CurrentValue: false,
						Frequency:    50,
						Labels:       []string{"radical"},
					},
				}},
				Layout: v1.Layout{
					Title:       "The Important App",
					Description: "???",
					Sections: []v1.Section{
						{
							Title:       "The performance indicator",
							Description: "Pay attention!",
							Indicators:  []string{"performance-indicator"},
						},
					},
					Owner: "Waldo",
				},
			},
		}))

	})
}
