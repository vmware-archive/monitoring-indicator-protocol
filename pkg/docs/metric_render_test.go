package docs_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cf-indicators/pkg/docs"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
)

func TestRenderMetricHTML(t *testing.T) {
	g := NewGomegaWithT(t)

	m := indicator.Metric{
		Title:       "Demo Latency",
		Name:        "latency",
		SourceID:    "demo id",
		Origin:      "demo origin",
		Description: "test description *bold text*",
		Type:        "gauge",
		Frequency:   "35s",
	}

	html, err := docs.MetricToHTML(m)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(html).To(ContainSubstring(`### <a id="demo-latency"></a>Demo Latency`))
	g.Expect(html).To(ContainSubstring(`<tbody><tr><th colspan="2" style="text-align: center;"><br> latency<br><br></th></tr>`))
	g.Expect(html).To(ContainSubstring("<p>test description <em>bold text</em></p>"))
	g.Expect(html).To(ContainSubstring(`<span><strong>Firehose Origin</strong>: demo origin</span>`))
	g.Expect(html).To(ContainSubstring(`<span><strong>Log Cache Source ID</strong>: demo id</span>`))
	g.Expect(html).To(ContainSubstring(`<span><strong>Type</strong>: gauge</span>`))
	g.Expect(html).To(ContainSubstring(`<span><strong>Frequency</strong>: 35s</span>`))
	g.Expect(html).ToNot(ContainSubstring("%%"))
}
