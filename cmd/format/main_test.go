package main_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"bytes"
	"code.cloudfoundry.org/indicators/pkg/go_test"
	"os"
	"os/exec"
)

func TestFormatBinary(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("accepts indicator yml file as a command line argument and returns formatted HTML", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath, "-format","bookbinder","../../example.yml")

		buffer := bytes.NewBuffer(nil)

		sess, err := gexec.Start(cmd, buffer, os.Stderr)
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

			g.Expect(html).To(ContainSubstring(`avg_over_time(demo_latency{source_id="doc",deployment="$deployment"}[5m])`))
		})

		t.Run("It does not have multiple % signs", func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(html).ToNot(ContainSubstring("%%"))
		})
	})

	t.Run("accepts indicator yml and returns grafana dashboards", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath, "-format","grafana","../../example.yml")

		buffer := bytes.NewBuffer(nil)

		sess, err := gexec.Start(cmd, buffer, os.Stderr)
		g.Expect(err).ToNot(HaveOccurred())

		g.Eventually(sess).Should(gexec.Exit(0))

		text := buffer.String()

		t.Run("it outputs indicators titles", func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(text).To(ContainSubstring(`"title":"doc_performance_indicator"`))
			g.Expect(text).To(ContainSubstring(`"expr":"avg_over_time(demo_latency{source_id=\"doc\",deployment=\"my-service-deployment\"}[5m])"`))
		})
	})
}
