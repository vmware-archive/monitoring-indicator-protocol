package indicator_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func TestValidDocument(t *testing.T) {
	t.Run("validation returns no errors if the document is valid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion:     "v0",
			Product: indicator.Product{Name: "valid", Version: "0.1.1"},
			Indicators: []indicator.Indicator{{
				Name:   "test_performance_indicator",
				PromQL: "prom",
			}},
			Documentation: indicator.Documentation{
				Title:       "Monitoring Test Product",
				Owner:       "Test Owner Team",
				Description: "Test description",
				Sections: []indicator.Section{{
					Title:       "Test Section",
					Description: "This section includes indicators and metrics",
				}},
			},
		}

		es := indicator.Validate(document)

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

		es := indicator.Validate(document)

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

		es := indicator.Validate(document)

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

		es := indicator.Validate(document)

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

		es := indicator.Validate(document)

		g.Expect(es).To(ConsistOf(
			errors.New("only apiVersion v0 is supported"),
		))
	})
}

func TestIndicatorValidation(t *testing.T) {

	t.Run("validation returns errors if any indicator field is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Indicators: []indicator.Indicator{
				{
					Name:   " ",
					PromQL: " ",
				},
			},
		}

		es := indicator.Validate(document)

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
					Name:   "not.valid",
					PromQL: " ",
				},
			},
		}

		es := indicator.Validate(document)

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
					Name:   `valid{labels="nope"}`,
					PromQL: `valid{labels="yep"}`,
				},
			},
		}

		es := indicator.Validate(document)

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
		))
	})
}
