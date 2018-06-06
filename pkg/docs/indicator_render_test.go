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
		MetricRefs:  []indicator.MetricRef{{Name: "latency", SourceID: "demo"}},
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
	g.Expect(html).To(ContainSubstring(`<tr><th colspan="2" style="text-align: center;"><br/> latency<br/><br/></th></tr>`))
	g.Expect(html).To(ContainSubstring("<p><em>test description</em> of kpi</p>"))
	g.Expect(html).To(ContainSubstring(`<td><code>avg_over_time(test_latency{source_id="test"}[100m])</code></td>`))
	g.Expect(html).To(ContainSubstring("<p><em>test response</em> of kpi</p>"))
	g.Expect(html).To(ContainSubstring("<p>Average over 100 minutes</p>"))
	g.Expect(html).To(ContainSubstring("<em>Red critical</em>: &gt; 1000<br/>"))
	g.Expect(html).To(ContainSubstring("<em>Yellow warning</em>: Dynamic<br/>"))
	g.Expect(html).To(ContainSubstring("<em>super_green</em>: &lt; 10<br/>"))
	g.Expect(html).ToNot(ContainSubstring("%%"))
}
