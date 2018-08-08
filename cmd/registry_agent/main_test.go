package main_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"os/exec"
	"os"
	"code.cloudfoundry.org/cf-indicators/pkg/vgo_test"
	"net/http"
	"github.com/onsi/gomega/ghttp"
)

func TestIndicatorRegistryAgent(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := vgo_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it sends an indicator document to the registry on an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryServer := ghttp.NewServer()
		defer registryServer.Close()

		registryServer.AppendHandlers(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})

		println(registryServer.URL())

		cmd := exec.Command(binPath,
			"--indicators-path", "./test_fixtures/",
			"--registry", registryServer.URL(),
			"--deployment", "abc-123",
			"--interval", "50ms")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()

		g.Eventually(registryServer.ReceivedRequests).Should(HaveLen(2))
	})
}
