package registry_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"net/http"
	"time"
	"code.cloudfoundry.org/cf-indicators/pkg/registry"
)

func TestRegistryAgent(t *testing.T) {
	t.Run("it sends an indicator document to the registry on an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryServer := ghttp.NewServer()
		defer registryServer.Close()

		registryServer.AppendHandlers(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})

		agent := registry.Agent{
			IndicatorsPath: "./test_fixtures/",
			RegistryURI:    registryServer.URL(),
			DeploymentName: "abc-123",
			IntervalTime:   50 * time.Millisecond}

		go agent.Start()

		g.Eventually(registryServer.ReceivedRequests).Should(HaveLen(2))
	})
}
