package v1_test

import (
	"errors"
	"io/ioutil"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	. "github.com/onsi/gomega"
)

func TestValidDocument(t *testing.T) {
	t.Run("validation returns no errors if the document is valid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
				Kind: "IndicatorDocument",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"new-metadata-value": "blah", "another-new-metadata-value": "blah2"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "valid", Version: "0.1.1"},
				Indicators: []v1.IndicatorSpec{{
					Name:   "test_performance_indicator",
					PromQL: "prom",
					Thresholds: []v1.Threshold{{
						Level:    "critical",
						Operator: v1.GreaterThan,
						Value:    0,
					}, {
						Level:    "warning",
						Operator: v1.GreaterThanOrEqualTo,
						Value:    10,
					}, {
						Level:    "warning",
						Operator: v1.LessThan,
						Value:    70,
					}, {
						Level:    "critical",
						Operator: v1.LessThanOrEqualTo,
						Value:    0,
					}, {
						Level:    "warning",
						Operator: v1.EqualTo,
						Value:    0,
					}, {
						Level:    "critical",
						Operator: v1.NotEqualTo,
						Value:    1000,
					}},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
				Layout: v1.Layout{
					Title:       "Monitoring Test Product",
					Owner:       "Test Owner Team",
					Description: "Test description",
					Sections: []v1.Section{{
						Title:       "Test Section",
						Description: "This section includes indicators and metrics",
						Indicators:  []string{"test_performance_indicator"},
					}},
				},
			},
		}

		es := document.Validate()

		g.Expect(es).To(BeEmpty())
	})
}

func TestProduct(t *testing.T) {
	t.Run("validation returns errors if product is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "", Version: "0.0.1"},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.spec.product.name in body should be at least 1 chars long"),
		))
	})
}

func TestVersion(t *testing.T) {
	t.Run("validation returns errors if version is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "product", Version: ""},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.spec.product.version in body should be at least 1 chars long"),
		))
	})
}

func TestAPIVersion(t *testing.T) {
	t.Run("validation returns errors if APIVersion is absent", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(errors.New("IndicatorDocument.apiVersion in body is required")))
	})

	t.Run("validation returns errors if APIVersion is not supported", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "fake-version",
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.apiVersion in body should be one of [indicatorprotocol.io/v1]"),
		))
	})
}

func TestSpec(t *testing.T) {
	t.Run("validation returns errors if spec doesn't exist", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.spec.product.name in body should be at least 1 chars long")))
		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.spec.product.version in body should be at least 1 chars long")))
	})
}

func TestKind(t *testing.T) {
	t.Run("should error in kind other than IndicatorDocument", func(t *testing.T) {
	    g := NewGomegaWithT(t)
		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				Kind: "notindicatordocument",
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.kind in body should be one of [IndicatorDocument]"),
		))
	})
}

func TestIndicator(t *testing.T) {

	t.Run("validation returns errors if any indicator field is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{
					{
						Name:         " ",
						PromQL:       "",
						Presentation: test_fixtures.DefaultPresentation(),
					},
				},
			},
		}

		es := document.Validate()

		g.Expect(es).To(And(
			ContainElement(errors.New("IndicatorDocument.spec.indicators.name in body should match '[a-zA-Z_:][a-zA-Z0-9_:]*'")),
			ContainElement(errors.New("IndicatorDocument.spec.indicators.promql in body should be at least 1 chars long")),
		))
	})

	t.Run("validation returns errors if indicator name is not valid promql", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{
					{
						Name:         "not.valid",
						PromQL:       "",
						Presentation: test_fixtures.DefaultPresentation(),
					},
				},
			},
		}

		es := document.Validate()

		g.Expect(es).To(And(
			ContainElement(errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)")),
			ContainElement(errors.New("IndicatorDocument.spec.indicators.promql in body should be at least 1 chars long")),
		))
	})

	t.Run("validation returns errors if indicator name is not valid promql", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{
					{
						Name:         `valid{labels="nope"}`,
						PromQL:       `valid{labels="yep"}`,
						Presentation: test_fixtures.DefaultPresentation(),
					},
				},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
		))
	})
}

func TestLayout(t *testing.T) {
	t.Run("validation returns error if layout contains non-existent indicator name", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"new-metadata-value": "blah", "another-new-metadata-value": "blah2"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "valid", Version: "0.1.1"},
				Indicators: []v1.IndicatorSpec{{
					Name:         "test_performance_indicator",
					PromQL:       "prom",
					Presentation: test_fixtures.DefaultPresentation(),
				}},
				Layout: v1.Layout{
					Title:       "Monitoring Test Product",
					Owner:       "Test Owner Team",
					Description: "Test description",
					Sections: []v1.Section{{
						Title:       "Test Section",
						Description: "This section includes indicators and metrics",
						Indicators:  []string{"test_performance_indicator", "cats"},
					}},
				},
			},
		}
		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("layout sections[0] indicators[1] references a non-existent indicator"),
		))
	})
}

func TestMetadata(t *testing.T) {
	t.Run("validation returns errors if metadata key is step", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"step": "my-step"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"),
		))
	})

	t.Run("validation returns errors if metadata key is step containing uppercase letters", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"StEp": "my-step"},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"),
		))
	})
}

func TestThreshold(t *testing.T) {
	t.Run("validation returns errors if threshold value is missing in v0 document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V0,
			},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{{
					Name:   "my_fair_indicator",
					PromQL: "rate(speech)",
					Thresholds: []v1.Threshold{{
						Level:    "warning",
						Operator: v1.UndefinedOperator,
						Value:    0,
					}},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.spec.indicators.thresholds.operator in body should be one of [lt lte gt gte eq neq]"),
		))
	})

	t.Run("validation returns errors if threshold operator is missing in v1 document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{{
					Name:   "my_fair_indicator",
					PromQL: "rate(speech)",
					Thresholds: []v1.Threshold{{
						Level:    "warning",
						Operator: v1.UndefinedOperator,
						Value:    0,
					}},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.spec.indicators.thresholds.operator in body should be one of [lt lte gt gte eq neq]"),
		))
	})
}

func TestChartType(t *testing.T) {
	t.Run("validation returns errors if chart type is invalid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1.IndicatorDocument{
			TypeMeta: metav1.TypeMeta{
				APIVersion: api_versions.V1,
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1.IndicatorSpec{{
					Name:   "my_fair_indicator",
					PromQL: "rate(speech)",
					Presentation: v1.Presentation{
						ChartType: "fakey-fake",
					},
				}},
			},
		}

		es := document.Validate()

		g.Expect(es).To(ContainElement(
			errors.New("IndicatorDocument.spec.indicators.presentation.chartType in body should be one of [step bar status quota]"),
		))
	})
}

func TestIndicatorDocumentSchema(t *testing.T) {
	t.Run("Accepts the example indicator document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		exampleDocBytes, err := ioutil.ReadFile("../../../../../example_indicators.yml")
		g.Expect(err).To(BeNil())

		_, ok := v1.ValidateBytesBySchema(exampleDocBytes, "IndicatorDocument")
		g.Expect(ok).To(BeTrue())
	})

	t.Run("Does not accept invalid indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)

		exampleDocBytes := []byte(`yaml: invalidindicator`)
		_, ok := v1.ValidateBytesBySchema(exampleDocBytes, "IndicatorDocument")
		g.Expect(ok).To(BeFalse())
	})
}
