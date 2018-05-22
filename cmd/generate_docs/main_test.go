package main_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"os/exec"
	"bytes"
	"os"
	"strings"
)

func TestGenerateDocsBinary(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := gexec.Build("./main.go")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("accepts indicator yml files as a command line argument and returns formatted HTML", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath, "./test_fixtures/valid.yml")

		buffer := bytes.NewBuffer(nil)

		sess, err := gexec.Start(cmd, buffer, os.Stderr)
		g.Expect(err).ToNot(HaveOccurred())

		g.Eventually(sess).Should(gexec.Exit(0))

		output := strings.Join(strings.Fields(buffer.String()),"")

		g.Expect(output).To(ContainSubstring(`<td>avg_over_time(demo_latency{source_id="test"}[5m])</td>`))
	})

	gexec.CleanupBuildArtifacts()
}
