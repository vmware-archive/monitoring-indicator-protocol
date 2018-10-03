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
			APIVersion: "v0",
			Product:    "valid",
			Version:    "0.1.1",
			Indicators: []indicator.Indicator{{
				Name:        "test_performance_indicator",
				PromQL:      "prom",
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
			Product: "",
			Version: "0.1.1",
		}

		es := indicator.Validate(document)

		g.Expect(es).To(ConsistOf(
			errors.New("product is required"),
		))
	})
}

func TestVersion(t *testing.T) {
	t.Run("validation returns errors if version is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			APIVersion: "v0",
			Product: "product",
			Version: "",
		}

		es := indicator.Validate(document)

		g.Expect(es).To(ConsistOf(
			errors.New("version is required"),
		))
	})
}

func TestAPIVersion(t *testing.T) {
	t.Run("validation returns errors if APIVersion is absent", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := indicator.Document{
			Product: "product",
			Version: "1",
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
			Product: "product",
			Version: "1",
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
			Product: "valid",
			Version: "1",
			Indicators: []indicator.Indicator{
				{
					Name:        " ",
					PromQL:      " ",
				},
			},
		}

		es := indicator.Validate(document)

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name is required"),
			errors.New("indicators[0] promql is required"),
		))
	})
}
