package indicator_test

import (
	"testing"

	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

func TestUpdateMetadata(t *testing.T) {
	t.Run("it replaces promql $EXPR with metadata tags", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadFile("test_fixtures/doc_value_interpolation.yml")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(BeEquivalentTo(v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IndicatorDocument",
				APIVersion: api_versions.V1,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "document name",
				Labels: map[string]string{
					"deployment":  "well-performing-deployment",
					"some_number": "450",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{
					{
						Name:   "test_performance_indicator",
						PromQL: `query_metric{source_id="well-performing-deployment"}`,
						Presentation: v1.Presentation{
							CurrentValue: false,
							ChartType:    v1.StepChart,
							Frequency:    0,
							Labels:       []string{},
							Units:        "short",
						},
						Thresholds: []v1.Threshold{
							{
								Level:    "critical",
								Operator: v1.GreaterThanOrEqualTo,
								Value:    450,
								Alert:    test_fixtures.DefaultAlert(),
							},
						},
					},
				},
				Layout: v1.Layout{
					Title: "well-performing-component - 0.0.1",
					Sections: []v1.Section{{
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

		g.Expect(d).To(Equal(v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
				Kind:       "IndicatorDocument",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "document name",
				Labels: map[string]string{"deployment": "well-performing-deployment"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{
					{
						Name:   "test_performance_indicator",
						PromQL: `query_metric{source_id="$deployment"}`,
						Presentation: v1.Presentation{
							CurrentValue: false,
							ChartType:    v1.StepChart,
							Frequency:    0,
							Labels:       []string{},
							Units:        "short",
						},
					},
				},
				Layout: v1.Layout{
					Title: "well-performing-component - 0.0.1",
					Sections: []v1.Section{{
						Title:      "Metrics",
						Indicators: []string{"test_performance_indicator"},
					}},
				},
			},
		}))
	})
}
