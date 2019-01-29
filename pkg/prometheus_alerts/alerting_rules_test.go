package prometheus_alerts_test

import (
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/prometheus_alerts"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAlertGeneration(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it makes prometheus alerts from the thresholds", func(t *testing.T) {
		g = NewGomegaWithT(t)
		document := indicator.Document{
			Indicators: []indicator.Indicator{
				{
					Thresholds: []indicator.Threshold{{}, {}},
				},
				{
					Thresholds: []indicator.Threshold{{}},
				},
			},
		}

		alertDoc := prometheus_alerts.AlertDocumentFrom(document)

		g.Expect(alertDoc.Groups[0].Rules).To(HaveLen(3))
	})

	t.Run("it generates a promql statement for less than statements", func(t *testing.T) {
		g = NewGomegaWithT(t)

		exprFor := func(op indicator.OperatorType) string {
			doc := indicator.Document{Indicators: []indicator.Indicator{
				{
					PromQL: `metric{source_id="fake-source"}`,
					Thresholds: []indicator.Threshold{{
						Level:    "warning",
						Operator: op,
						Value:    0.99999999999999, //keep many nines to ensure we don't blow float parsing to 1
					}},
				}},
			}

			return getFirstRule(doc).Expr
		}

		g.Expect(exprFor(indicator.LessThanOrEqualTo)).To(Equal(`metric{source_id="fake-source"} <= 0.99999999999999`))
		g.Expect(exprFor(indicator.LessThan)).To(Equal(`metric{source_id="fake-source"} < 0.99999999999999`))
		g.Expect(exprFor(indicator.EqualTo)).To(Equal(`metric{source_id="fake-source"} == 0.99999999999999`))
		g.Expect(exprFor(indicator.NotEqualTo)).To(Equal(`metric{source_id="fake-source"} != 0.99999999999999`))
		g.Expect(exprFor(indicator.GreaterThan)).To(Equal(`metric{source_id="fake-source"} > 0.99999999999999`))
		g.Expect(exprFor(indicator.GreaterThanOrEqualTo)).To(Equal(`metric{source_id="fake-source"} >= 0.99999999999999`))
	})

	t.Run("sets the name to the indicator's name", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := indicator.Document{
			Indicators: []indicator.Indicator{{
				Name:       "indicator_lol",
				Thresholds: []indicator.Threshold{{}},
			}},
		}

		g.Expect(getFirstRule(doc).Alert).To(Equal("indicator_lol"))
	})

	t.Run("sets the labels", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := indicator.Document{
			Product:  indicator.Product{Name: "product-lol", Version: "beta.9"},
			Metadata: map[string]string{"meta-lol": "data-lol"},
			Indicators: []indicator.Indicator{{
				Name:         "indicator_lol",
				Thresholds: []indicator.Threshold{{
					Level: "warning",
				}},
			}},
		}

		g.Expect(getFirstRule(doc).Labels).To(Equal(map[string]string{
			"product":                 "product-lol",
			"version":                 "beta.9",
			"level":                   "warning",
			"meta-lol":                "data-lol",
		}))
	})

	t.Run("sets the annotations to the documentation block", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := indicator.Document{
			Indicators: []indicator.Indicator{{
				Documentation: map[string]string{"title-lol": "Indicator LOL"},
				Thresholds:    []indicator.Threshold{{}},
			}},
		}

		g.Expect(getFirstRule(doc).Annotations).To(Equal(map[string]string{
			"title-lol": "Indicator LOL",
		}))
	})
}

func getFirstRule(from indicator.Document) prometheus_alerts.Rule {
	return prometheus_alerts.AlertDocumentFrom(from).Groups[0].Rules[0]
}
