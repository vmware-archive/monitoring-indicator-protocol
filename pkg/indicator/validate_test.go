package indicator_test

import (
	"errors"
	"testing"

	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func TestValidDocument(t *testing.T) {
	t.Run("validation returns no errors if the document is valid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "valid", Version: "0.1.1"},
			Metadata:   map[string]string{"new-metadata-value": "blah", "another-new-metadata-value": "blah2"},
			Indicators: []indicator.Indicator{{
				Name:   "test_performance_indicator",
				PromQL: "prom",
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.GreaterThan,
					Value:    0,
				}, {
					Level:    "warning",
					Operator: indicator.GreaterThanOrEqualTo,
					Value:    10,
				}, {
					Level:    "warning",
					Operator: indicator.LessThan,
					Value:    70,
				}, {
					Level:    "critical",
					Operator: indicator.LessThanOrEqualTo,
					Value:    0,
				}, {
					Level:    "warning",
					Operator: indicator.EqualTo,
					Value:    0,
				}, {
					Level:    "critical",
					Operator: indicator.NotEqualTo,
					Value:    1000,
				}},
				Presentation: test_fixtures.DefaultPresentation(),
			}},
			Layout: indicator.Layout{
				Title:       "Monitoring Test Product",
				Owner:       "Test Owner Team",
				Description: "Test description",
				Sections: []indicator.Section{{
					Title:       "Test Section",
					Description: "This section includes indicators and metrics",
				}},
			},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(BeEmpty())
	})
}

func TestProduct(t *testing.T) {
	t.Run("validation returns errors if product is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "", Version: "0.0.1"},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("product name is required"),
		))
	})
}

func TestVersion(t *testing.T) {
	t.Run("validation returns errors if version is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "product", Version: ""},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("product version is required"),
		))
	})
}

func TestAPIVersion(t *testing.T) {
	t.Run("validation returns errors if APIVersion is absent", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			Product: indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("apiVersion is required"),
			errors.New("only apiVersion v0 is supported"),
		))
	})

	t.Run("validation returns errors if APIVersion is not v0", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "fake-version",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("only apiVersion v0 is supported"),
		))
	})
}

func TestIndicator(t *testing.T) {

	t.Run("validation returns errors if any indicator field is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Indicators: []indicator.Indicator{
				{
					Name:         " ",
					PromQL:       " ",
					Presentation: test_fixtures.DefaultPresentation(),
				},
			},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name is required"),
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
			errors.New("indicators[0] promql is required"),
		))
	})

	t.Run("validation returns errors if indicator name is not valid promql", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Indicators: []indicator.Indicator{
				{
					Name:         "not.valid",
					PromQL:       " ",
					Presentation: test_fixtures.DefaultPresentation(),
				},
			},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
			errors.New("indicators[0] promql is required"),
		))
	})

	t.Run("validation returns errors if indicator name is not valid promql", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Indicators: []indicator.Indicator{
				{
					Name:         `valid{labels="nope"}`,
					PromQL:       `valid{labels="yep"}`,
					Presentation: test_fixtures.DefaultPresentation(),
				},
			},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
		))
	})
}

func TestMetadata(t *testing.T) {
	t.Run("validation returns errors if metadata key is step", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Metadata:   map[string]string{"step": "my-step"},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"),
		))
	})
}

func TestThreshold(t *testing.T) {
	t.Run("validation returns errors if threshold value is missing", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Indicators: []indicator.Indicator{{
				Name:   "my_fair_indicator",
				PromQL: "rate(speech)",
				Thresholds: []indicator.Threshold{{
					Level:    "warning",
					Operator: indicator.Undefined,
					Value:    0,
				}},
				Presentation: test_fixtures.DefaultPresentation(),
			}},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0].thresholds[0] value is required, one of [lt, lte, eq, neq, gte, gt] must be provided as a float"),
		))
	})
}

func TestChartType(t *testing.T) {
	t.Run("validation returns errors if chart type is invalid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Indicators: []indicator.Indicator{{
				Name:   "my_fair_indicator",
				PromQL: "rate(speech)",
				Presentation: &indicator.Presentation{
					ChartType: "fakey-fake",
				},
			}},
		}

		es := indicator.ValidateForRegistry(document)

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] invalid chartType provided: 'fakey-fake' - valid chart types are [step bar status]"),
		))
	})
}
