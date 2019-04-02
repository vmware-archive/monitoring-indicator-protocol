package prometheus_alerts_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
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
				Name: "indicator_lol",
				Thresholds: []indicator.Threshold{{
					Level: "warning",
				}},
			}},
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

	t.Run("sets the alert for", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := indicator.Document{
			Product:  indicator.Product{Name: "product-lol", Version: "beta.9"},
			Metadata: map[string]string{"meta-lol": "data-lol"},
			Indicators: []indicator.Indicator{{
				Name:       "indicator_lol",
				PromQL:     "promql_expression",
				Thresholds: []indicator.Threshold{{}},
				Alert: indicator.Alert{
					For: "40h",
				},
			}},
		}

		g.Expect(getFirstRule(doc).For).To(Equal("40h"))
	})

	t.Run("interpolates $step", func(t *testing.T) {
		g = NewGomegaWithT(t)

		doc := indicator.Document{
			Product:  indicator.Product{Name: "product-lol", Version: "beta.9"},
			Metadata: map[string]string{"meta-lol": "data-lol"},
			Indicators: []indicator.Indicator{{
				Name:       "indicator_lol",
				PromQL:     "super_query(promql_expression[$step])[$step]",
				Thresholds: []indicator.Threshold{{}},
				Alert: indicator.Alert{
					Step: "12m",
				},
			}},
		}

		g.Expect(getFirstRule(doc).Expr).To(Equal("super_query(promql_expression[12m])[12m] < 0"))
	})

	t.Run("creates a filename based on product name and contents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		document := indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "test_product",
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: `test_query{deployment="test_deployment"}`,
				Alert:  test_fixtures.DefaultAlert(),
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.LessThan,
					Value:    5,
				}},
				Presentation:  test_fixtures.DefaultPresentation(),
				Documentation: map[string]string{"title": "Test Indicator Title"},
			}},
			Layout: indicator.Layout{
				Title: "Test Dashboard",
				Sections: []indicator.Section{
					{
						Title: "Test Section Title",
					},
				},
			},
		}
		document.Layout.Sections[0].Indicators = document.Indicators

		docBytes, err := json.Marshal(document)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(prometheus_alerts.AlertDocumentFilename(docBytes, "test_product")).To(Equal("test_product_c47308287f60c776ba39f27aaa743bca7c6c6387.yml"))
	})
}

func getFirstRule(from indicator.Document) prometheus_alerts.Rule {
	return prometheus_alerts.AlertDocumentFrom(from).Groups[0].Rules[0]
}
