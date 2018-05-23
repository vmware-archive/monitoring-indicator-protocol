package main_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"os/exec"
	"bytes"
	"os"
)

func TestGenerateDocsBinary(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := gexec.Build("./main.go")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("accepts indicator yml file as a command line argument and returns formatted HTML", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath, "./test_fixtures/valid.yml")

		buffer := bytes.NewBuffer(nil)

		sess, err := gexec.Start(cmd, buffer, os.Stderr)
		g.Expect(err).ToNot(HaveOccurred())

		g.Eventually(sess).Should(gexec.Exit(0))

		html := buffer.String()

		t.Run("It displays document title and description", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`<h1 class="title-container">Monitoring Test Product</h1>`))
			g.Expect(html).To(ContainSubstring(`<p>Test description</p>`))
		})

		t.Run("It displays indicator sections", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`<h2 id="key-performance-indicators">Key Performance Indicators</h2>`))
			g.Expect(html).To(ContainSubstring(`<p>This section includes indicators</p>`))

			g.Expect(html).To(ContainSubstring(`<h3 id="test-performance-indicator">Test Performance Indicator</h3>`))
		})

		t.Run("It displays metric sections", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`<h2 id="other-metrics-available">Other Metrics Available</h2>`))
			g.Expect(html).To(ContainSubstring(`<p>This section includes metrics</p>`))

			g.Expect(html).To(ContainSubstring(`<h3 id="demo-latency">Demo Latency</h3>`))
		})

		t.Run("It does not have multiple % signs", func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(html).ToNot(ContainSubstring("%%"))
		})
	})

	gexec.CleanupBuildArtifacts()
}
