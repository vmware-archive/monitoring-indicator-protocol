package main_test

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal/indicator-protocol/pkg/go_test"
)

func TestFormatBinary(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("complains if no indicators file path specified", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath)

		buffer := bytes.NewBuffer(nil)

		sess, _ := gexec.Start(cmd, buffer, buffer)
		g.Eventually(sess).Should(gexec.Exit(1))
		g.Expect(buffer.String()).To(ContainSubstring("-indicators flag is required"))
	})

	t.Run("outputs formatted HTML", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath,
			"-format", "html",
			"-metadata", "deployment=my-other-service-deployment",
			"-indicators", "../../example_indicators.yml")

		buffer := bytes.NewBuffer(nil)

		sess, err := gexec.Start(cmd, buffer, buffer)
		g.Expect(err).ToNot(HaveOccurred())

		g.Eventually(sess).Should(gexec.Exit(0))

		html := buffer.String()

		t.Run("It displays document title and description", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`<title>Monitoring Document Product</title>`))
			g.Expect(html).To(ContainSubstring(`<h1>Monitoring Document Product</h1>`))
			g.Expect(html).To(ContainSubstring(`Document description`))
		})

		t.Run("It displays indicator sections", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`<h2><a id="indicators"></a>Indicators</h2>`))
			g.Expect(html).To(ContainSubstring(`This section includes indicators`))

			g.Expect(html).To(ContainSubstring(`<h3><a id="doc_performance_indicator"></a>Doc Performance Indicator</h3>`))

			g.Expect(html).To(ContainSubstring(`avg_over_time(demo_latency{source_id="$source_id",deployment="$deployment"}[5m])`))
		})

		t.Run("It does not have multiple % signs", func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(html).ToNot(ContainSubstring("%%"))
		})
	})

	t.Run("outputs bookbinder formatted HTML", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath,
			"-format", "bookbinder",
			"-metadata", "deployment=my-other-service-deployment",
			"-indicators", "../../example_indicators.yml")

		buffer := bytes.NewBuffer(nil)

		sess, err := gexec.Start(cmd, buffer, buffer)
		g.Expect(err).ToNot(HaveOccurred())

		g.Eventually(sess).Should(gexec.Exit(0))

		html := buffer.String()

		t.Run("It displays document title and description", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`title: Monitoring Document Product`))
			g.Expect(html).To(ContainSubstring(`Document description`))
		})

		t.Run("It displays indicator sections", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`## <a id="indicators"></a>Indicators`))
			g.Expect(html).To(ContainSubstring(`This section includes indicators`))

			g.Expect(html).To(ContainSubstring(`### <a id="doc_performance_indicator"></a>Doc Performance Indicator`))

			g.Expect(html).To(ContainSubstring(`avg_over_time(demo_latency{source_id="$source_id",deployment="$deployment"}[5m])`))
		})

		t.Run("It does not have multiple % signs", func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(html).ToNot(ContainSubstring("%%"))
		})
	})

	t.Run("outputs prometheus alert configuration", func(t *testing.T) {
		t.Run("with no metadata flag", func(t *testing.T) {
			g := NewGomegaWithT(t)

			cmd := exec.Command(binPath,
				"-format", "prometheus-alerts",
				"-indicators", "../../example_indicators.yml")

			buffer := bytes.NewBuffer(nil)

			sess, err := gexec.Start(cmd, buffer, buffer)
			g.Expect(err).ToNot(HaveOccurred())

			g.Eventually(sess).Should(gexec.Exit(0))

			prometheusAlertConfigurationYML := buffer.String()

			fileBytes, err := ioutil.ReadFile("test_fixtures/prometheus_alert.yml")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(prometheusAlertConfigurationYML).To(MatchYAML(fileBytes))
		})

		t.Run("with metadata flag", func(t *testing.T) {
			g := NewGomegaWithT(t)

			cmd := exec.Command(binPath,
				"-format", "prometheus-alerts",
				"-metadata", "deployment=my-other-service-deployment",
				"-indicators", "../../example_indicators.yml")

			buffer := bytes.NewBuffer(nil)

			sess, err := gexec.Start(cmd, buffer, buffer)
			g.Expect(err).ToNot(HaveOccurred())

			g.Eventually(sess).Should(gexec.Exit(0))

			prometheusAlertConfigurationYML := buffer.String()

			g.Expect(prometheusAlertConfigurationYML).To(ContainSubstring(`deployment: my-other-service-deployment`))
		})
	})

	t.Run("outputs grafana dashboards", func(t *testing.T) {
		t.Run("with no metadata flag", func(t *testing.T) {
			g := NewGomegaWithT(t)

			cmd := exec.Command(binPath,
				"-format", "grafana",
				"-indicators", "../../example_indicators.yml")

			buffer := bytes.NewBuffer(nil)

			sess, err := gexec.Start(cmd, buffer, buffer)
			g.Expect(err).ToNot(HaveOccurred())

			g.Eventually(sess).Should(gexec.Exit(0))

			text := buffer.String()

			t.Run("it outputs indicators titles", func(t *testing.T) {
				g := NewGomegaWithT(t)
				g.Expect(text).To(ContainSubstring(`"title": "doc_performance_indicator"`))
				g.Expect(text).To(ContainSubstring(`"expr": "avg_over_time(demo_latency{source_id=\"my-metric-source\",deployment=\"my-service-deployment\"}[5m])"`))
			})
		})

		t.Run("with metadata flag", func(t *testing.T) {
			g := NewGomegaWithT(t)

			cmd := exec.Command(binPath,
				"-format", "grafana",
				"-metadata", "deployment=my-other-service-deployment",
				"-indicators", "../../example_indicators.yml")

			buffer := bytes.NewBuffer(nil)

			sess, err := gexec.Start(cmd, buffer, buffer)
			g.Expect(err).ToNot(HaveOccurred())

			g.Eventually(sess).Should(gexec.Exit(0))

			text := buffer.String()

			t.Run("it outputs indicators titles", func(t *testing.T) {
				g := NewGomegaWithT(t)
				g.Expect(text).To(ContainSubstring(`"title": "doc_performance_indicator"`))
				g.Expect(text).To(ContainSubstring(`"expr": "avg_over_time(demo_latency{source_id=\"my-metric-source\",deployment=\"my-other-service-deployment\"}[5m])"`))
			})
		})
	})
}
