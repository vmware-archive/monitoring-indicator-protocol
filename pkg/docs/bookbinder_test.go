package docs_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/docs"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

func TestRenderDocumentHTML(t *testing.T) {
	g := NewGomegaWithT(t)

	indicators := []v1alpha1.IndicatorSpec{
		{
			Name: "test_indicator",
			Documentation: map[string]string{

				"title":       "Test Indicator",
				"description": "*test description* of kpi",
				"response":    "*test response* of kpi",
				"measurement": "Average over 100 minutes",
			},
			PromQL: `avg_over_time(test_latency{source_id="demo_source"}[100m])`,
			Thresholds: []v1alpha1.Threshold{
				{
					Level:    "warning",
					Operator: v1alpha1.GreaterThan,
					Value:    500,
				},
				{
					Level:    "critical",
					Operator: v1alpha1.GreaterThan,
					Value:    1000,
				},
			},
		},
	}
	document := v1alpha1.IndicatorDocument{
		Spec: v1alpha1.IndicatorDocumentSpec{
		Indicators: indicators,
		Layout: v1alpha1.Layout{
			Title:       "Test Document",
			Owner:       "Test Owner",
			Description: "This is a document for testing `code`",
			Sections: []v1alpha1.Section{
				{
					Title:       "Test Indicators Section",
					Description: "This is a section of indicator documentation for testing `other code`",
					Indicators:  []string{"test_indicator"},
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
