package docs_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cf-indicators/pkg/docs"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
)

func TestRenderIndicatorHTML(t *testing.T) {
	g := NewGomegaWithT(t)

	indicator := indicator.Indicator{
		Name:        "test_indicator",
		Title:       "Test Indicator",
		Description: "*test description* of kpi",
		Response:    "*test response* of kpi",
		PromQL:      `avg_over_time(test_latency{source_id="test"}[100m])`,
		Metrics:  []indicator.Metric{{
			Title:       "Demo Latency Metric",
			Origin:      "origin1",
			SourceID:    "demo",
			Name:        "latency",
			Type:        "metric_type",
			Description: "This is a metric",
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
			{
				Level:    "super_green",
				Operator: indicator.LessThan,
				Value:    10,
			},
		},
	}

	html, err := docs.IndicatorToHTML(indicator)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(html).To(ContainSubstring(`### <a id="test-indicator"></a>Test Indicator`))
	g.Expect(html).To(ContainSubstring("<p><em>test description</em> of kpi</p>"))
	g.Expect(html).To(ContainSubstring(`<code>avg_over_time(test_latency{source_id="test"}[100m])</code>`))

	t.Run("should contain metrics table", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(html).To(ContainSubstring(`<strong>latency</strong>`))
		g.Expect(html).To(ContainSubstring(`This is a metric`))
		g.Expect(html).To(ContainSubstring(`<strong>firehose origin</strong>: origin1`))
		g.Expect(html).To(ContainSubstring(`<strong>log-cache source_id</strong>: demo`))
		g.Expect(html).To(ContainSubstring(`<strong>type</strong>: metric_type`))
	})

	g.Expect(html).To(ContainSubstring("<p><em>test response</em> of kpi</p>"))
	g.Expect(html).To(ContainSubstring("<p>Average over 100 minutes</p>"))
	g.Expect(html).To(ContainSubstring("<em>Red critical</em>: &gt; 1000<br/>"))
	g.Expect(html).To(ContainSubstring("<em>Yellow warning</em>: Dynamic<br/>"))
	g.Expect(html).To(ContainSubstring("<em>super_green</em>: &lt; 10<br/>"))
	g.Expect(html).ToNot(ContainSubstring("%%"))
}
