package indicator_test

import (
  . "github.com/onsi/gomega"
  "testing"

  "code.cloudfoundry.org/indicators/pkg/indicator"
)

func TestReturnsCompleteDocument(t *testing.T) {
  g := NewGomegaWithT(t)
  d, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product: well-performing-component
version: 0.0.1
metadata:
  deployment: well-performing-deployment

indicators:
- name: test_performance_indicator
  documentation:
    title: Test Performance Indicator
    description: This is a valid markdown description.
    recommended_response: Panic!
    threshold_note: Threshold Note Text
  promql: prom
  thresholds:
  - level: warning
    gte: 50

documentation:
  title: Monitoring Test Product
  description: Test description
  sections:
  - title: Test Section
    description: This section includes indicators and metrics
    indicators:
    - test_performance_indicator
`))
  g.Expect(err).ToNot(HaveOccurred())

  g.Expect(d).To(Equal(indicator.Document{
    APIVersion: "v0",
    Product:    "well-performing-component",
    Version:    "0.0.1",
    Metadata:   map[string]string{"deployment": "well-performing-deployment"},
    Indicators: []indicator.Indicator{
      {
        Name:   "test_performance_indicator",
        PromQL: "prom",
        Thresholds: []indicator.Threshold{
          {
            Level:    "warning",
            Operator: indicator.GreaterThanOrEqualTo,
            Value:    50,
          },
        },
        Documentation: map[string]string{
          "title":                "Test Performance Indicator",
          "description":          "This is a valid markdown description.",
          "recommended_response": "Panic!",
          "threshold_note":       "Threshold Note Text",
        },
      },
    },
    Documentation: indicator.Documentation{
      Title:       "Monitoring Test Product",
      Description: "Test description",
      Sections: []indicator.Section{{
        Title:       "Test Section",
        Description: "This section includes indicators and metrics",
        Indicators: []indicator.Indicator{{
          Name:   "test_performance_indicator",
          PromQL: "prom",
          Thresholds: []indicator.Threshold{
            {
              Level:    "warning",
              Operator: indicator.GreaterThanOrEqualTo,
              Value:    50,
            },
          },
          Documentation: map[string]string{
            "title":                "Test Performance Indicator",
            "description":          "This is a valid markdown description.",
            "recommended_response": "Panic!",
            "threshold_note":       "Threshold Note Text",
          },
        }},
      }},
    },
  }))
}

func TestReturnsAnEmptyListWhenNoIndicatorsArePassed(t *testing.T) {
  g := NewGomegaWithT(t)

  d, err := indicator.ReadIndicatorDocument([]byte(`---
indicators: []`))
  g.Expect(err).ToNot(HaveOccurred())

  g.Expect(d.Indicators).To(HaveLen(0))
}

func TestReturnsAConvertedIndicator(t *testing.T) {
  g := NewGomegaWithT(t)

  d, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
- name: test-kpi
  promql: prom
  thresholds:
  - lt: 0
    level: warning
  - lte: 1.2
    level: warning
  - eq: 0.2
    level: warning
  - neq: 123
    level: warning
  - gte: 642
    level: warning
  - gt: 1.222225
    level: warning`))

  g.Expect(err).ToNot(HaveOccurred())

  g.Expect(d.Indicators).To(Equal([]indicator.Indicator{{
    Name:   "test-kpi",
    PromQL: "prom",
    Thresholds: []indicator.Threshold{
      {
        Level:    "warning",
        Operator: indicator.LessThan,
        Value:    0,
      },
      {
        Level:    "warning",
        Operator: indicator.LessThanOrEqualTo,
        Value:    1.2,
      },
      {
        Level:    "warning",
        Operator: indicator.EqualTo,
        Value:    0.2,
      },
      {
        Level:    "warning",
        Operator: indicator.NotEqualTo,
        Value:    123,
      },
      {
        Level:    "warning",
        Operator: indicator.GreaterThanOrEqualTo,
        Value:    642,
      },
      {
        Level:    "warning",
        Operator: indicator.GreaterThan,
        Value:    1.222225,
      },
    },
  }}))
}

func TestReturnsAnErrorIfTheYAMLIsUnparsable(t *testing.T) {
  g := NewGomegaWithT(t)

  _, err := indicator.ReadIndicatorDocument([]byte(`--`))
  g.Expect(err).To(HaveOccurred())
}

func TestReturnsAnErrorIfAThresholdHasNoValue(t *testing.T) {
  g := NewGomegaWithT(t)

  _, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - level: warning
  `))
  g.Expect(err).To(HaveOccurred())
}

func TestReturnsAnErrorIfAThresholdHasABadFloatValue(t *testing.T) {
  g := NewGomegaWithT(t)

  _, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - gte: abs
    level: warning
  `))
  g.Expect(err).To(HaveOccurred())
}

func TestReturnsErrors(t *testing.T) {
  t.Run("if section references non-existent indicator", func(t *testing.T) {
    g := NewGomegaWithT(t)

    _, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product: my-product
version: 1
indicators: []
documentation:
  title: docs
  description: desc
  sections:
  - title: metric section
    description: metric desc
    indicators:
    - not_found
  `))
    g.Expect(err).To(MatchError(ContainSubstring("documentation.sections[0].indicators[0] references non-existent indicator")))
  })
}
