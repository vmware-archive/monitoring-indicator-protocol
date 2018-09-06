package registry_test

import (
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"testing"

	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"code.cloudfoundry.org/cf-indicators/pkg/registry"
	"net/http"
	"time"
)

func TestRegistryAgent(t *testing.T) {
	t.Run("it sends an indicator document to the registry on an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryServer := ghttp.NewServer()
		defer registryServer.Close()

		receivedDocument := make(chan indicator.Document, 1)

		registryServer.AppendHandlers(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			content, err := ioutil.ReadAll(r.Body)
			g.Expect(err).ToNot(HaveOccurred())

			document, err := indicator.ReadIndicatorDocument(content)
			g.Expect(err).To(Not(HaveOccurred()))

			receivedDocument <- document

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})

		document, err := ioutil.ReadFile("./test_fixtures/indicators.yml")
		g.Expect(err).To(Not(HaveOccurred()))

		agent := registry.Agent{
			IndicatorsDocuments: [][]byte{document},
			RegistryURI:         registryServer.URL(),
			DeploymentName:      "abc-123",
			IntervalTime:        50 * time.Millisecond}

		go agent.Start()

		g.Eventually(registryServer.ReceivedRequests).Should(HaveLen(2))

		request := registryServer.ReceivedRequests()[0]
		queryParams := request.URL.Query()
		g.Expect(queryParams.Get("deployment")).To(Equal("abc-123"))

		g.Expect((<-receivedDocument).Labels["product"]).To(Equal("product-name"))
	})
}
