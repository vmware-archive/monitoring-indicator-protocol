package registry_test

import (
	. "github.com/onsi/gomega"
	"testing"

	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/indicators/pkg/go_test"
	"code.cloudfoundry.org/indicators/pkg/registry"
)

func TestAPIClient_IndicatorDocuments(t *testing.T) {
	g := NewGomegaWithT(t)
	exampleJSON, e := ioutil.ReadFile("../../pkg/registry/test_fixtures/example_response.json")
	g.Expect(e).ToNot(HaveOccurred())
	http.HandleFunc("/v1/indicator-documents", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(exampleJSON)
	})

	server := http.Server{
		Addr: "localhost:8080",
	}

	go server.ListenAndServe()

	defer server.Close()
	go_test.WaitForHTTPServer("localhost:8080", time.Second)

	t.Run("it parses the indicator response into Document Structs", func(t *testing.T) {
		g := NewGomegaWithT(t)
		c := registry.NewAPIClient("http://localhost:8080", http.DefaultClient)

		documents, e := c.IndicatorDocuments()
		g.Expect(e).ToNot(HaveOccurred())
		g.Expect(documents[0].Product).To(Equal("my-component"))
		g.Expect(documents[0].Version).To(Equal("1.2.3"))
	})
}
