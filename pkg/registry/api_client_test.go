package registry_test

import (
	. "github.com/onsi/gomega"
	"testing"
	"time"

	"net/http"

	"code.cloudfoundry.org/indicators/pkg/go_test"
	"code.cloudfoundry.org/indicators/pkg/registry"
)

func TestAPIClient_IndicatorDocuments(t *testing.T) {
	c := registry.NewAPIClient("http://localhost:8080", http.DefaultClient)

	t.Run("it fetches the payload on the /v1/indicator-documents endpoint", func(t *testing.T) {
		g := NewGomegaWithT(t)

		http.HandleFunc("/v1/indicator-documents", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("payload"))
		})

		server := http.Server{
			Addr: "localhost:8080",
		}

		go server.ListenAndServe()
		defer server.Close()

		go_test.WaitForHTTPServer("localhost:8080", time.Second)

		g.Expect(c.IndicatorDocuments()).To(Equal(registry.IndicatorDocuments("payload")))
	})

	t.Run("it returns an error if the the client get fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := c.IndicatorDocuments()
		g.Expect(err).To(HaveOccurred())
	})
}
