package docs_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/docs"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func TestRenderIndicatorHTML(t *testing.T) {
	t.Run("it renders indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicator := indicator.Indicator{
			Name: "test_indicator",
			Documentation: map[string]string{
				"title":                "Test Indicator",
				"description":          "*test description* of kpi",
				"recommended_response": "*test response* of kpi",
				"measurement":          "Average over 100 minutes",
				"threshold_note":       "dynamic!",
			},
			PromQL: `avg_over_time(test_latency{source_id="test"}[100m])`,

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
				{
					Level:    "super_green",
					Operator: indicator.LessThan,
					Value:    10,
				},
			},
		}

		ind := docs.NewIndicatorPresenter(indicator)
		html := string(ind.HTML())

		g.Expect(html).To(ContainSubstring("<p><em>test description</em> of kpi</p>"))
		g.Expect(html).To(ContainSubstring(`<code>avg_over_time(test_latency{source_id="test"}[100m])</code>`))

		g.Expect(html).To(ContainSubstring("<p><em>test response</em> of kpi</p>"))
		g.Expect(html).To(ContainSubstring("<p>Average over 100 minutes</p>"))
		g.Expect(html).To(ContainSubstring("<em>Red critical</em>: &gt; 1000<br/>"))
		g.Expect(html).To(ContainSubstring("<em>Yellow warning</em>: &gt; 500<br/>"))
		g.Expect(html).To(ContainSubstring("<em>super_green</em>: &lt; 10<br/>"))
		g.Expect(html).To(ContainSubstring("dynamic!"))
		g.Expect(html).ToNot(ContainSubstring("Threshold Note"))
		g.Expect(html).To(ContainSubstring("Recommended Response"))

		g.Expect(html).ToNot(ContainSubstring("%%"))
	})

	t.Run("it handles camelcase for documentation fields", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicator := indicator.Indicator{
			Name: "test_indicator",
			Documentation: map[string]string{
				"title":                "Test Indicator",
				"description":          "*test description* of kpi",
				"recommendedResponse": "*test response* of kpi",
				"thresholdNote":       "dynamic!",
			},
			PromQL: `avg_over_time(test_latency{source_id="test"}[100m])`,

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
				{
					Level:    "super_green",
					Operator: indicator.LessThan,
					Value:    10,
				},
			},
		}

		ind := docs.NewIndicatorPresenter(indicator)
		html := string(ind.HTML())

		g.Expect(html).To(ContainSubstring("dynamic!"))
		g.Expect(html).ToNot(ContainSubstring("Threshold Note"))
		g.Expect(html).To(ContainSubstring("Recommended Response"))
		g.Expect(html).To(ContainSubstring("<p><em>test response</em> of kpi</p>"))

		g.Expect(html).ToNot(ContainSubstring("%%"))
	})

	t.Run("it skips the threshold section when absent", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicator := indicator.Indicator{
			Name: "test_indicator",
			Documentation: map[string]string{
				"title":                "Test Indicator",
				"description":          "*test description* of kpi",
				"recommendedResponse": "*test response* of kpi",
				"measurement":          "Average over 100 minutes",
			},
			PromQL: `avg_over_time(test_latency{source_id="test"}[100m])`,

			Thresholds: []indicator.Threshold{},
		}

		ind := docs.NewIndicatorPresenter(indicator)
		html := string(ind.HTML())

		g.Expect(html).ToNot(ContainSubstring("<th>Thresholds</th>"))
	})
}
