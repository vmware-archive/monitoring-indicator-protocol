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
		Name:        "demo.latency",
		Description: "test description *bold text*",
	}

	html, err := docs.MetricToHTML(m)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(html).To(ContainSubstring(`<h3 id="demo-latency">Demo Latency</h3>`))
	g.Expect(html).To(ContainSubstring(`<tbody><tr><th colspan="2" style="text-align: center;"><br> demo.latency<br><br></th></tr>`))
	g.Expect(html).To(ContainSubstring("<p>test description <em>bold text</em></p>"))
	g.Expect(html).ToNot(ContainSubstring("%%"))
}
