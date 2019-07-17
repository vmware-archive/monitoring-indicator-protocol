package indicator_test

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

func TestUpdateMetadata(t *testing.T) {
	t.Run("it replaces promql $EXPR with metadata tags", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadFile("test_fixtures/doc.yml")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(BeEquivalentTo(v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				Kind: "IndicatorDocument",
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"deployment": "well-performing-deployment"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Name:   "test_performance_indicator",
						PromQL: `query_metric{source_id="well-performing-deployment"}`,
						Alert: v1alpha1.Alert{
							For:  "1m",
							Step: "1m",
						},
						Presentation: v1alpha1.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
					},
				},
				Layout: v1alpha1.Layout{
					Sections: []v1alpha1.Section{{
						Title:      "Metrics",
						Indicators: []string{"test_performance_indicator"},
					}},
				},
			},
		}))
	})

	t.Run("it does not replaces promql $EXPR with metadata tags when passed flag", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadFile("test_fixtures/doc.yml", indicator.SkipMetadataInterpolation)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(Equal(v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
				Kind: "IndicatorDocument",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"deployment": "well-performing-deployment"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Name:   "test_performance_indicator",
						PromQL: `query_metric{source_id="$deployment"}`,
						Alert: v1alpha1.Alert{
							For:  "1m",
							Step: "1m",
						},
						Presentation: v1alpha1.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
					},
				},
				Layout: v1alpha1.Layout{
					Sections: []v1alpha1.Section{{
						Title:      "Metrics",
						Indicators: []string{"test_performance_indicator"},
					}},
				},
			},
		}))
	})
}
