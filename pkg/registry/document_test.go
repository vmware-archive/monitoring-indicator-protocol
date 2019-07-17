package registry_test

import (
	"testing"

	. "github.com/onsi/gomega"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestDocumentTranslation(t *testing.T) {
	t.Run("it translates", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicatorDoc := registry.APIV0Document{
			APIVersion: "apps.pivotal.io/v1alpha1",
			Metadata: map[string]string{
				"someKey": "someValue",
			},

			Product: registry.APIV0Product{
				Name:    "important-application",
				Version: "1.0",
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

		g.Expect(registry.ToIndicatorDocument(indicatorDoc)).To(Equal(v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				Kind:       "IndicatorDocument",
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"someKey": "someValue",
				},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "important-application",
					Version: "1.0",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "performance-indicator",
					PromQL: "someQuery",
					Thresholds: []v1alpha1.Threshold{{
						Level:    "warning",
						Operator: v1alpha1.LessThanOrEqualTo,
						Value:    100,
					}},
					Alert: v1alpha1.Alert{
						For:  "30s",
						Step: "5s",
					},
					Documentation: map[string]string{
						"anotherKey": "anotherValue",
					},
					Presentation: v1alpha1.Presentation{
						ChartType:    "bar",
						CurrentValue: false,
						Frequency:    50,
						Labels:       []string{"radical"},
					},
				}},
				Layout: v1alpha1.Layout{
					Title:       "The Important App",
					Description: "???",
					Sections: []v1alpha1.Section{
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
