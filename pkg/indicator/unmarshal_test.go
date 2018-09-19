package indicator_test

import (
	"code.cloudfoundry.org/indicators/pkg/indicator"

	"testing"

	. "github.com/onsi/gomega"
)

func TestReturnsCompleteDocument(t *testing.T) {
	g := NewGomegaWithT(t)
	d, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
labels:
  product: well-performing-component
metrics:
- name: latency
  source_id: demo
  origin: demo
  title: Demo Latency
  type: metricType
  frequency: 60s
  description: A test metric for testing

indicators:
- name: test_performance_indicator
  title: Test Performance Indicator
  metrics:
  - name: latency
    source_id: demo
  measurement: Measurement Text
  promql: prom
  thresholds:
  - level: warning
    gte: 50
    dynamic: true
  description: This is a valid markdown description.
  response: Panic!

documentation:
  title: Monitoring Test Product
  description: Test description
  sections:
  - title: Test Section
    description: This section includes indicators and metrics
    indicators:
    - name: test_performance_indicator
    metrics:
    - name: latency
      source_id: demo
`))
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(d).To(Equal(indicator.Document{
		APIVersion: "v0",
		Labels: map[string]string{"product":"well-performing-component"},
		Metrics: []indicator.Metric{
			{
				Title:       "Demo Latency",
				Origin:      "demo",
				SourceID:    "demo",
				Name:        "latency",
				Type:        "metricType",
				Description: "A test metric for testing",
				Frequency:   "60s",
			},
		},
		Indicators: []indicator.Indicator{
			{
				Name:        "test_performance_indicator",
				Title:       "Test Performance Indicator",
				Description: "This is a valid markdown description.",
				PromQL:      "prom",
				Thresholds: []indicator.Threshold{
					{
						Level:    "warning",
						Dynamic:  true,
						Operator: indicator.GreaterThanOrEqualTo,
						Value:    50,
					},
				},
				Metrics: []indicator.Metric{{
					Title:       "Demo Latency",
					Origin:      "demo",
					SourceID:    "demo",
					Name:        "latency",
					Type:        "metricType",
					Description: "A test metric for testing",
					Frequency:   "60s",
				}},
				Response:    "Panic!",
				Measurement: "Measurement Text",
			},
		},
		Documentation: indicator.Documentation{
			Title:       "Monitoring Test Product",
			Description: "Test description",
			Sections: []indicator.Section{{
				Title:       "Test Section",
				Description: "This section includes indicators and metrics",
				Indicators: []indicator.Indicator{{
					Name:        "test_performance_indicator",
					Title:       "Test Performance Indicator",
					Description: "This is a valid markdown description.",
					PromQL:      "prom",
					Thresholds: []indicator.Threshold{
						{
							Level:    "warning",
							Dynamic:  true,
							Operator: indicator.GreaterThanOrEqualTo,
							Value:    50,
						},
					},
					Metrics: []indicator.Metric{{
						Title:       "Demo Latency",
						Origin:      "demo",
						SourceID:    "demo",
						Name:        "latency",
						Type:        "metricType",
						Description: "A test metric for testing",
						Frequency:   "60s",
					}},
					Response:    "Panic!",
					Measurement: "Measurement Text",
				}},
				Metrics: []indicator.Metric{{
					Title:       "Demo Latency",
					Origin:      "demo",
					SourceID:    "demo",
					Name:        "latency",
					Type:        "metricType",
					Description: "A test metric for testing",
					Frequency:   "60s",
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

func TestReturnsAConvertedMetric(t *testing.T) {
	g := NewGomegaWithT(t)

	metricYAML := `---
metrics:
- name: latency
  source_id: demo
  origin: demo
  title: Demo Latency
  description: A test metric for testing`

	indicatorDocument, err := indicator.ReadIndicatorDocument([]byte(metricYAML))
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(indicatorDocument.Metrics).To(ContainElement(indicator.Metric{
		Title:       "Demo Latency",
		Name:        "latency",
		SourceID:    "demo",
		Origin:      "demo",
		Description: "A test metric for testing",
	}))
}

func TestReturnsAConvertedIndicator(t *testing.T) {
	g := NewGomegaWithT(t)

	d, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
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
	t.Run("if indicator references non-existent metric", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds: []
  metrics:
  - name: not_found
    source_id: not_found_source
  `))
		g.Expect(err).To(MatchError(ContainSubstring("indicators[0].metrics[0] references non-existent metric")))
	})

	t.Run("if section references non-existent metric", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := indicator.ReadIndicatorDocument([]byte(`---
indicators: []
metrics: []
documentation:
  title: docs
  description: desc
  sections:
  - title: metric section
    description: metric desc
    metrics:
    - name: not_found
      source_id: not_found_source
  `))
		g.Expect(err).To(MatchError(ContainSubstring("documentation.sections[0].metrics[0] references non-existent metric")))
	})

	t.Run("if section references non-existent indicator", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := indicator.ReadIndicatorDocument([]byte(`---
indicators: []
metrics: []
documentation:
  title: docs
  description: desc
  sections:
  - title: metric section
    description: metric desc
    indicators:
    - name: not_found
  `))
		g.Expect(err).To(MatchError(ContainSubstring("documentation.sections[0].indicators[0] references non-existent indicator")))
	})
}
