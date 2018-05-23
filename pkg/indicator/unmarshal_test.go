package indicator_test

import (
	"github.com/cloudfoundry-incubator/event-producer/pkg/indicator"

	"testing"
	. "github.com/onsi/gomega"
)

func TestReturnsAnEmptyListWhenNoIndicatorsArePassed(t *testing.T) {
	g := NewGomegaWithT(t)

	d, err := indicator.ReadIndicatorDocument([]byte(`---
performance_indicators: []`))
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(d.Indicators).To(HaveLen(0))
}

func TestReturnsAConvertedMetric(t *testing.T) {
	g := NewGomegaWithT(t)

	metricYAML := `---
metrics:
- name: demo.latency
  title: Demo Latency
  description: A test metric for testing
  type: gauge
  unit: milliseconds`

	indicatorDocument, err := indicator.ReadIndicatorDocument([]byte(metricYAML))
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(indicatorDocument.Metrics).To(ContainElement(indicator.Metric{
		Title:       "Demo Latency",
		Name:        "demo.latency",
		Description: "A test metric for testing",
	}))
}

func TestReturnsAConvertedIndicator(t *testing.T) {
	g := NewGomegaWithT(t)

	d, err := indicator.ReadIndicatorDocument([]byte(`---
performance_indicators:
- name: test-kpi
  description: desc
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
    dynamic: false
  - gte: 642
    level: warning
    dynamic: true
  - gt: 1.222225
    level: warning`))

	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(d.Indicators).To(Equal([]indicator.Indicator{{
		Name:        "test-kpi",
		Description: "desc",
		PromQL:      "prom",
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
				Dynamic:  false,
				Operator: indicator.NotEqualTo,
				Value:    123,
			},
			{
				Level:    "warning",
				Dynamic:  true,
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
performance_indicators:
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
performance_indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - gte: abs
    level: warning
  `))
	g.Expect(err).To(HaveOccurred())
}
