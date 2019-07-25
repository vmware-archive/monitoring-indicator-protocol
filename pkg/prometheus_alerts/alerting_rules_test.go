package prometheus_alerts_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestAlertGeneration(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it makes prometheus alerts from the thresholds", func(t *testing.T) {
		g = NewGomegaWithT(t)
		document := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Thresholds: []v1alpha1.Threshold{{}, {}},
					},
					{
						Thresholds: []v1alpha1.Threshold{{}},
					},
				},
			},
		}

		alertDoc := prometheus_alerts.AlertDocumentFrom(document)

		g.Expect(alertDoc.Groups[0].Rules).To(HaveLen(3))
	})

	t.Run("it generates a promql statement for less than statements", func(t *testing.T) {
		g = NewGomegaWithT(t)

		exprFor := func(op v1alpha1.ThresholdOperator) string {
			doc := v1alpha1.IndicatorDocument{
				Spec: v1alpha1.IndicatorDocumentSpec{
					Indicators: []v1alpha1.IndicatorSpec{
						{
							PromQL: `metric{source_id="fake-source"}`,
							Thresholds: []v1alpha1.Threshold{{
								Level:    "warning",
								Operator: op,
								Value:    0.99999999999999, //keep many nines to ensure we don't blow float parsing to 1
							}},
						},
					},
				},
			}

			return getFirstRule(doc).Expr
		}

		g.Expect(exprFor(v1alpha1.LessThanOrEqualTo)).To(Equal(`metric{source_id="fake-source"} <= 0.99999999999999`))
		g.Expect(exprFor(v1alpha1.LessThan)).To(Equal(`metric{source_id="fake-source"} < 0.99999999999999`))
		g.Expect(exprFor(v1alpha1.EqualTo)).To(Equal(`metric{source_id="fake-source"} == 0.99999999999999`))
		g.Expect(exprFor(v1alpha1.NotEqualTo)).To(Equal(`metric{source_id="fake-source"} != 0.99999999999999`))
		g.Expect(exprFor(v1alpha1.GreaterThan)).To(Equal(`metric{source_id="fake-source"} > 0.99999999999999`))
		g.Expect(exprFor(v1alpha1.GreaterThanOrEqualTo)).To(Equal(`metric{source_id="fake-source"} >= 0.99999999999999`))
	})

	t.Run("sets the name to the indicator's name", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:       "indicator_lol",
					Thresholds: []v1alpha1.Threshold{{}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Alert).To(Equal("indicator_lol"))
	})

	t.Run("sets the labels", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1alpha1.IndicatorDocument{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"meta-lol": "data-lol"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "product-lol", Version: "beta.9"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name: "indicator_lol",
					Thresholds: []v1alpha1.Threshold{{
						Level: "warning",
					}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Labels).To(Equal(map[string]string{
			"product":  "product-lol",
			"version":  "beta.9",
			"level":    "warning",
			"meta-lol": "data-lol",
		}))
	})

	t.Run("sets the annotations to the documentation block", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Indicators: []v1alpha1.IndicatorSpec{{
					Documentation: map[string]string{"title-lol": "Indicator LOL"},
					Thresholds:    []v1alpha1.Threshold{{}},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Annotations).To(Equal(map[string]string{
			"title-lol": "Indicator LOL",
		}))
	})

	t.Run("sets the alert for", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1alpha1.IndicatorDocument{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"meta-lol": "data-lol"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "product-lol", Version: "beta.9"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:       "indicator_lol",
					PromQL:     "promql_expression",
					Thresholds: []v1alpha1.Threshold{{}},
					Alert: v1alpha1.Alert{
						For: "40h",
					},
				}},
			},
		}

		g.Expect(getFirstRule(doc).For).To(Equal("40h"))
	})

	t.Run("interpolates $step", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"meta-lol": "data-lol"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "product-lol", Version: "beta.9"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "indicator_lol",
					PromQL: "super_query(promql_expression[$step])[$step]",
					Thresholds: []v1alpha1.Threshold{{
						Level:    "warning",
						Operator: v1alpha1.LessThan,
						Value:    0,
					}},
					Alert: v1alpha1.Alert{
						Step: "12m",
					},
				}},
			},
		}

		g.Expect(getFirstRule(doc).Expr).To(Equal("super_query(promql_expression[12m])[12m] < 0"))
	})

	t.Run("creates a filename based on product name and document contents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Alert:  test_fixtures.DefaultAlert(),
					Thresholds: []v1alpha1.Threshold{{
						Level:    "critical",
						Operator: v1alpha1.LessThan,
						Value:    5,
					}},
					Presentation:  test_fixtures.DefaultPresentation(),
					Documentation: map[string]string{"title": "Test Indicator Title"},
				}},
				Layout: v1alpha1.Layout{
					Title: "Test Dashboard",
					Sections: []v1alpha1.Section{
						{
							Title:      "Test Section Title",
							Indicators: []string{"test_indicator"},
						},
					},
				},
			},
		}

		docBytes, err := json.Marshal(document)

		g.Expect(err).ToNot(HaveOccurred())
		filename := prometheus_alerts.AlertDocumentFilename(docBytes, "test_product")
		g.Expect(filename).To(MatchRegexp("test_product_[0-9a-f]{40}.yml"))
	})
}

func getFirstRule(from v1alpha1.IndicatorDocument) prometheus_alerts.Rule {
	return prometheus_alerts.AlertDocumentFrom(from).Groups[0].Rules[0]
}
