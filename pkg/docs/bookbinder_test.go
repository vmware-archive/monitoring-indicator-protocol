package docs_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/docs"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func TestRenderDocumentHTML(t *testing.T) {
	g := NewGomegaWithT(t)

	document := indicator.Document{
		Layout: indicator.Layout{
			Title:       "Test Document",
			Owner:       "Test Owner",
			Description: "This is a document for testing `code`",
			Sections: []indicator.Section{
				{
					Title:       "Test Indicators Section",
					Description: "This is a section of indicator documentation for testing `other code`",
					Indicators: []indicator.Indicator{
						{
							Name: "test_indicator",
							Documentation: map[string]string{

								"title":       "Test Indicator",
								"description": "*test description* of kpi",
								"response":    "*test response* of kpi",
								"measurement": "Average over 100 minutes",
							},
							PromQL: `avg_over_time(test_latency{source_id="demo_source"}[100m])`,
							Thresholds: []indicator.Threshold{
								{
									Level:    "warning",
									Operator: indicator.GreaterThan,
									Value:    500,
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
			},
		},
	}

	html, err := docs.DocumentToBookbinder(document)
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

		g.Expect(html).To(ContainSubstring(`### <a id="test_indicator"></a>Test Indicator`))
	})
}
