package main_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"bytes"
	"os"
	"os/exec"
	"code.cloudfoundry.org/indicators/pkg/go_test"
)

func TestGenerateDocsBinary(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
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
			g.Expect(html).To(ContainSubstring(`title: Monitoring Test Product`))
			g.Expect(html).To(ContainSubstring(`Test description`))
		})

		t.Run("It displays indicator sections", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`## <a id="key-performance-indicators"></a>Key Performance Indicators`))
			g.Expect(html).To(ContainSubstring(`This section includes indicators`))

			g.Expect(html).To(ContainSubstring(`### <a id="test-performance-indicator"></a>Test Performance Indicator`))
		})

		t.Run("It displays metric sections", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(html).To(ContainSubstring(`## <a id="other-metrics-available"></a>Other Metrics Available`))
			g.Expect(html).To(ContainSubstring(`This section includes metrics`))

			g.Expect(html).To(ContainSubstring(`### <a id="demo-latency"></a>Demo Latency`))
		})

		t.Run("It does not have multiple % signs", func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(html).ToNot(ContainSubstring("%%"))
		})
	})
}
