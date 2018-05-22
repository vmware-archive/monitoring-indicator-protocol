package docs_test

import (
	"testing"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"
	"github.com/cloudfoundry-incubator/event-producer/pkg/docs"
)

func TestRenderHTML(t *testing.T) {
	g := NewGomegaWithT(t)

	indicator := kpi.KPI{
		Name:        "Test KPI",
		Description: "*test description* of kpi",
		Response:    "*test response* of kpi",
		PromQL:      `avg_over_time(test_latency{source_id="test"}[100m])`,
		Metrics:     []string{"test.latency"},
		Measurement: "Average over 100 minutes",
		Thresholds: []kpi.Threshold{{
			Level:    "critical",
			Operator: kpi.GreaterThan,
			Value:    1000,
		}},
	}

	html, err := docs.HTML(indicator)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(html).To(ContainSubstring(`<h3 id="test-kpi"><a id="TestKPI"></a>Test KPI</h3>`))
	g.Expect(html).To(ContainSubstring(`<tr><th colspan="2" style="text-align: center;"><br/> test.latency<br/><br/><br/></th></tr>`))
	g.Expect(html).To(ContainSubstring("<p><em>test description</em> of kpi</p>"))
	g.Expect(html).To(ContainSubstring(`<td>avg_over_time(test_latency{source_id="test"}[100m])</td>`))
	g.Expect(html).To(ContainSubstring("<p><em>test response</em> of kpi</p>"))
	g.Expect(html).To(ContainSubstring("<p>Average over 100 minutes</p>"))
	g.Expect(html).To(ContainSubstring("<em>Red critical</em>: &gt; 1000<br/>"))
}
