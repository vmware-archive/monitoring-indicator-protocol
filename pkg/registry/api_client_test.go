package registry_test

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestAPIClient_IndicatorDocuments(t *testing.T) {
	g := NewGomegaWithT(t)
	exampleJSON, e := ioutil.ReadFile("test_fixtures/example_response.json")
	g.Expect(e).ToNot(HaveOccurred())
	http.HandleFunc("/v1/indicator-documents", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(exampleJSON)
	})

	registerHits := make(chan struct{})
	http.HandleFunc("/v1/register", func(writer http.ResponseWriter, request *http.Request) {
		go func() { registerHits <- struct{}{} }()
		bytes, err := ioutil.ReadAll(request.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bytes).To(ContainSubstring("spooky test value"))
	})

	server := http.Server{
		Addr: "localhost:14892",
	}

	go func() {
		log.Print(server.ListenAndServe())
	}()
	defer server.Close()
	err := go_test.WaitForTCPServer("localhost:14892", 5*time.Second)
	g.Expect(err).ToNot(HaveOccurred())

	c := registry.NewAPIClient("http://localhost:14892", http.DefaultClient)
	t.Run("it parses the indicator response into Document Structs", func(t *testing.T) {
		g := NewGomegaWithT(t)

		documents, e := c.IndicatorDocuments()
		g.Expect(e).ToNot(HaveOccurred())
		g.Expect(documents[0].Spec.Product.Name).To(Equal("my-component"))
		g.Expect(documents[0].Spec.Product.Version).To(Equal("1.2.3"))
	})

	t.Run("it can post a document", func(t *testing.T) {
		g := NewGomegaWithT(t)
		err := c.Register(v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name: "spooky test value",
				},
			},
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Eventually(registerHits, 5 * time.Second).Should(Receive())
	})
}
