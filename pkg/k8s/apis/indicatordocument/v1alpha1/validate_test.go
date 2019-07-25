package v1alpha1_test

import (
	"errors"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	. "github.com/onsi/gomega"
)

func TestValidDocument(t *testing.T) {
	t.Run("validation returns no errors if the document is valid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"new-metadata-value": "blah", "another-new-metadata-value": "blah2"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "valid", Version: "0.1.1"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "test_performance_indicator",
					PromQL: "prom",
					Thresholds: []v1alpha1.Threshold{{
						Level:    "critical",
						Operator: v1alpha1.GreaterThan,
						Value:    0,
					}, {
						Level:    "warning",
						Operator: v1alpha1.GreaterThanOrEqualTo,
						Value:    10,
					}, {
						Level:    "warning",
						Operator: v1alpha1.LessThan,
						Value:    70,
					}, {
						Level:    "critical",
						Operator: v1alpha1.LessThanOrEqualTo,
						Value:    0,
					}, {
						Level:    "warning",
						Operator: v1alpha1.EqualTo,
						Value:    0,
					}, {
						Level:    "critical",
						Operator: v1alpha1.NotEqualTo,
						Value:    1000,
					}},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
				Layout: v1alpha1.Layout{
					Title:       "Monitoring Test Product",
					Owner:       "Test Owner Team",
					Description: "Test description",
					Sections: []v1alpha1.Section{{
						Title:       "Test Section",
						Description: "This section includes indicators and metrics",
						Indicators:  []string{"test_performance_indicator"},
					}},
				},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(BeEmpty())
	})
}

func TestProduct(t *testing.T) {
	t.Run("validation returns errors if product is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "", Version: "0.0.1"},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("product name is required"),
		))
	})
}

func TestVersion(t *testing.T) {
	t.Run("validation returns errors if version is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "product", Version: ""},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("product version is required"),
		))
	})
}

func TestAPIVersion(t *testing.T) {
	t.Run("validation returns errors if APIVersion is absent", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("apiVersion is required"),
			errors.New("invalid apiVersion, supported versions are: [apps.pivotal.io/v1alpha1]"),
		))
	})

	t.Run("validation returns errors if APIVersion is not supported", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "fake-version",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate("v0", "apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("invalid apiVersion, supported versions are: [v0 apps.pivotal.io/v1alpha1]"),
		))
	})
}

func TestIndicator(t *testing.T) {

	t.Run("validation returns errors if any indicator field is blank", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Name:         " ",
						PromQL:       " ",
						Presentation: test_fixtures.DefaultPresentation(),
					},
				},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name is required"),
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
			errors.New("indicators[0] promql is required"),
		))
	})

	t.Run("validation returns errors if indicator name is not valid promql", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Name:         "not.valid",
						PromQL:       " ",
						Presentation: test_fixtures.DefaultPresentation(),
					},
				},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
			errors.New("indicators[0] promql is required"),
		))
	})

	t.Run("validation returns errors if indicator name is not valid promql", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{
					{
						Name:         `valid{labels="nope"}`,
						PromQL:       `valid{labels="yep"}`,
						Presentation: test_fixtures.DefaultPresentation(),
					},
				},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] name must be valid promql with no labels (see https://prometheus.io/docs/practices/naming)"),
		))
	})
}

func TestLayout(t *testing.T) {
	t.Run("validation returns error if layout contains non-existent indicator name", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"new-metadata-value": "blah", "another-new-metadata-value": "blah2"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "valid", Version: "0.1.1"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:         "test_performance_indicator",
					PromQL:       "prom",
					Presentation: test_fixtures.DefaultPresentation(),
				}},
				Layout: v1alpha1.Layout{
					Title:       "Monitoring Test Product",
					Owner:       "Test Owner Team",
					Description: "Test description",
					Sections: []v1alpha1.Section{{
						Title:       "Test Section",
						Description: "This section includes indicators and metrics",
						Indicators:  []string{"test_performance_indicator", "cats"},
					}},
				},
			},
		}
		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("layout sections[0] indicators[1] references a non-existent indicator"),
		))
	})
}

func TestMetadata(t *testing.T) {
	t.Run("validation returns errors if metadata key is step", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"step": "my-step"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"),
		))
	})

	t.Run("validation returns errors if metadata key is step containing uppercase letters", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"StEp": "my-step"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("metadata cannot contain `step` key (see https://github.com/pivotal/monitoring-indicator-protocol/wiki#metadata)"),
		))
	})
}

func TestThreshold(t *testing.T) {
	t.Run("validation returns errors if threshold value is missing in v0 document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v0",
			},
			ObjectMeta: v1.ObjectMeta{},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "my_fair_indicator",
					PromQL: "rate(speech)",
					Thresholds: []v1alpha1.Threshold{{
						Level:    "warning",
						Operator: v1alpha1.Undefined,
						Value:    0,
					}},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
			},
		}

		es := document.Validate("v0")

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0].thresholds[0] value is required, one of [lt, lte, eq, neq, gte, gt] must be provided as a float"),
		))
	})

	t.Run("validation returns errors if threshold operator is missing in v1alpha1 document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "my_fair_indicator",
					PromQL: "rate(speech)",
					Thresholds: []v1alpha1.Threshold{{
						Level:    "warning",
						Operator: v1alpha1.Undefined,
						Value:    0,
					}},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0].thresholds[0] operator [lt, lte, eq, neq, gte, gt] is required"),
		))
	})
}

func TestChartType(t *testing.T) {
	t.Run("validation returns errors if chart type is invalid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		document := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps.pivotal.io/v1alpha1",
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "well-performing-component", Version: "0.0.1"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "my_fair_indicator",
					PromQL: "rate(speech)",
					Presentation: v1alpha1.Presentation{
						ChartType: "fakey-fake",
					},
				}},
			},
		}

		es := document.Validate("apps.pivotal.io/v1alpha1")

		g.Expect(es).To(ConsistOf(
			errors.New("indicators[0] invalid chartType provided - valid chart types are [step bar status quota]"),
		))
	})
}
