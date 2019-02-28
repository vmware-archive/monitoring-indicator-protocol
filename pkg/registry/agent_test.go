package registry_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestRegistryAgent(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

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

		agent := registry.Agent{
			DocumentFinder: registry.DocumentFinder{Glob: "./test_fixtures/job-a/indicators.yml"},
			RegistryURI:    registryServer.URL(),
			IntervalTime:   50 * time.Millisecond,
			Client:         &http.Client{},
		}

		go agent.Start()

		g.Eventually(registryServer.ReceivedRequests).Should(HaveLen(2))

		document := <-receivedDocument
		g.Expect(document.Metadata["deployment"]).To(Equal("abc-123"))
		g.Expect(document.Product.Name).To(Equal("job-a-product"))
	})
}

func TestDocumentFinder(t *testing.T) {
	df := &registry.DocumentFinder{
		Glob: "./test_fixtures/job-a/indicators.yml",
	}

	t.Run("it returns documents matching the glob", func(t *testing.T) {
		g := NewGomegaWithT(t)

		documents, err := df.FindAll()

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(documents).To(HaveLen(1))
	})

	t.Run("it updates documents matching the glob when the glob list changes", func(t *testing.T) {
		g := NewGomegaWithT(t)

		df.Glob = "./test_fixtures/*/indicators.yml"

		documents, err := df.FindAll()

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(documents).To(HaveLen(2))
	})

}
