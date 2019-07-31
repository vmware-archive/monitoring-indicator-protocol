package registry_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"io/ioutil"
	"net/http"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestAPIClient_IndicatorDocuments(t *testing.T) {
	g := NewGomegaWithT(t)
	exampleJSON, e := ioutil.ReadFile("test_fixtures/example_response.json")
	g.Expect(e).ToNot(HaveOccurred())
	http.HandleFunc("/v1/indicator-documents", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(exampleJSON)
	})

	server := http.Server{
		Addr: "localhost:8975",
	}

	go server.ListenAndServe()

	defer server.Close()
	go_test.WaitForTCPServer("localhost:8975", time.Second)

	t.Run("it parses the indicator response into Document Structs", func(t *testing.T) {
		g := NewGomegaWithT(t)
		c := registry.NewAPIClient("http://localhost:8975", http.DefaultClient)

		documents, e := c.IndicatorDocuments()
		g.Expect(e).ToNot(HaveOccurred())
		g.Expect(documents[0].Spec.Product.Name).To(Equal("my-component"))
		g.Expect(documents[0].Spec.Product.Version).To(Equal("1.2.3"))
	})
}
