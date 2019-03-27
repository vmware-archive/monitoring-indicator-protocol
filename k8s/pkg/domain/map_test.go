package domain_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/domain"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMap(t *testing.T) {
	t.Run("works with a complete document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		threshold := float64(100)
		k8sDoc := v1alpha1.IndicatorDocument{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"level": "high"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "my-product",
					Version: "my-version",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "my-indicator",
					Promql: "my_promql",
					Alert: v1alpha1.Alert{
						For:  "10m",
						Step: "1m",
					},
					Thresholds: []v1alpha1.Threshold{{
						Level: "critical",
						Eq:    &threshold,
					}},
					Documentation: map[string]string{"docs": "explained"},
				}},
				Layout: v1alpha1.Layout{
					Owner:       "me",
					Title:       "my awesome indicators",
					Description: "enough said",
					Sections: []v1alpha1.Section{{
						Name:        "my section",
						Description: "the only section",
						Indicators:  []string{"my-indicator"},
					}},
				},
			},
		}

		i := indicator.Indicator{
			Name:   "my-indicator",
			PromQL: "my_promql",
			Thresholds: []indicator.Threshold{{
				Level:    "critical",
				Operator: indicator.EqualTo,
				Value:    100,
			}},
			Alert: indicator.Alert{
				For:  "10m",
				Step: "1m",
			},
			Documentation: map[string]string{"docs": "explained"},
		}

		domainDoc := indicator.Document{
			Product: indicator.Product{
				Name:    "my-product",
				Version: "my-version",
			},
			Metadata:   map[string]string{"level": "high"},
			Indicators: []indicator.Indicator{i},
			Layout: indicator.Layout{
				Title:       "my awesome indicators",
				Description: "enough said",
				Sections: []indicator.Section{{
					Title:       "my section",
					Description: "the only section",
					Indicators:  []indicator.Indicator{i},
				}},
				Owner: "me",
			},
		}
		g.Expect(domain.Map(&k8sDoc)).To(BeEquivalentTo(domainDoc))
	})

	t.Run("works with duplicate indicators in the layout", func(t *testing.T) {
		g := NewGomegaWithT(t)

		k8sDoc := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "my-product",
					Version: "my-version",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "my-indicator",
					Promql: "my_promql",
				}},
				Layout: v1alpha1.Layout{
					Owner:       "me",
					Title:       "my awesome indicators",
					Description: "enough said",
					Sections: []v1alpha1.Section{{
						Name:        "my section",
						Description: "the only section",
						Indicators:  []string{"my-indicator", "my-indicator"},
					}},
				},
			},
		}

		i := indicator.Indicator{
			Name:       "my-indicator",
			PromQL:     "my_promql",
			Thresholds: []indicator.Threshold{},
		}

		domainDoc := indicator.Document{
			Product: indicator.Product{
				Name:    "my-product",
				Version: "my-version",
			},
			Indicators: []indicator.Indicator{i},
			Layout: indicator.Layout{
				Title:       "my awesome indicators",
				Description: "enough said",
				Sections: []indicator.Section{{
					Title:       "my section",
					Description: "the only section",
					Indicators:  []indicator.Indicator{i, i},
				}},
				Owner: "me",
			},
		}
		g.Expect(domain.Map(&k8sDoc)).To(BeEquivalentTo(domainDoc))
	})
}
