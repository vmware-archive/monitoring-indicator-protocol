package docs_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cf-indicators/pkg/docs"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
)

func TestConvertIndicatorDocument(t *testing.T) {
	g := NewGomegaWithT(t)

	in := indicator.Document{
		Metrics: []indicator.Metric{
			{
				Title:       "Demo Latency",
				Name:        "latency",
				SourceID:    "demo",
				Description: "A test metric for testing",
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
				MetricRefs: []indicator.MetricRef{{
					Name:     "latency",
					SourceID: "demo",
				}},
				Response:    "Panic!",
				Measurement: "Measurement Text",
			},
		},
		Documentation: indicator.Documentation{
			Title:       "Monitoring Test Product",
			Owner:       "Test Owner Team",
			Description: "Test description",
			Sections: []indicator.Section{{
				Title:       "Test Section",
				Description: "This section includes indicators and metrics",
				IndicatorRefs: []indicator.IndicatorRef{{
					Name: "test_performance_indicator",
				}},
				MetricRefs: []indicator.MetricRef{{
					Name:     "latency",
					SourceID: "demo",
				}},
			}},
		},
	}

	g.Expect(docs.ConvertIndicatorDocument(in)).To(Equal(docs.Documentation{
		Title:       "Monitoring Test Product",
		Owner:       "Test Owner Team",
		Description: "Test description",
		Sections: []docs.Section{
			{
				Title:       "Test Section",
				Description: "This section includes indicators and metrics",
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
						MetricRefs: []indicator.MetricRef{{
							Name:     "latency",
							SourceID: "demo",
						}},
						Response:    "Panic!",
						Measurement: "Measurement Text",
					},
				},
				Metrics: []indicator.Metric{
					{
						Title:       "Demo Latency",
						Name:        "latency",
						SourceID:    "demo",
						Description: "A test metric for testing",
					},
				},
			},
		},
	}))
}

func TestRenderDocumentHTML(t *testing.T) {
	g := NewGomegaWithT(t)

	document := docs.Documentation{
		Title:       "Test Document",
		Owner:       "Test Owner",
		Description: "This is a document for testing `code`",
		Sections: []docs.Section{
			{
				Title:       "Test Indicators Section",
				Description: "This is a section of indicator documentation for testing `other code`",
				Indicators: []indicator.Indicator{
					{
						Name:        "test_indicator",
						Title:       "Test Indicator",
						Description: "*test description* of kpi",
						Response:    "*test response* of kpi",
						PromQL:      `avg_over_time(test_latency{source_id="test"}[100m])`,
						MetricRefs: []indicator.MetricRef{{
							Name:     "latency",
							SourceID: "demo",
						}},
						Measurement: "Average over 100 minutes",
						Thresholds: []indicator.Threshold{
							{
								Level:    "warning",
								Operator: indicator.GreaterThan,
								Value:    500,
								Dynamic:  true,
							},
							{
								Level:    "critical",
								Operator: indicator.GreaterThan,
								Value:    1000,
							},
						},
					},
				},
			},
			{
				Title:       "Test Metrics Section",
				Description: "This is a section of metric documentation for testing `yet more code`",
				Metrics: []indicator.Metric{
					{
						SourceID:    "test",
						Origin:      "test",
						Name:        "metric",
						Title:       "Test Metric",
						Description: "*test description* of metric",
					},
				},
			},
		},
	}

	html, err := docs.DocumentToHTML(document)
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It displays document title and description", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(html).To(ContainSubstring(`title: Test Document`))
		g.Expect(html).To(ContainSubstring(`<p>This is a document for testing <code>code</code></p>`))
	})

	t.Run("It displays indicator sections", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(html).To(ContainSubstring(`## <a id="test-indicators-section"></a>Test Indicators Section`))
		g.Expect(html).To(ContainSubstring(`<p>This is a section of indicator documentation for testing <code>other code</code></p>`))

		g.Expect(html).To(ContainSubstring(`### <a id="test-indicator"></a>Test Indicator`))
	})

	t.Run("It displays metric sections", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(html).To(ContainSubstring(`## <a id="test-metrics-section"></a>Test Metrics Section`))
		g.Expect(html).To(ContainSubstring(`<p>This is a section of metric documentation for testing <code>yet more code</code></p>`))

		g.Expect(html).To(ContainSubstring(`### <a id="test-metric"></a>Test Metric`))
	})
}
