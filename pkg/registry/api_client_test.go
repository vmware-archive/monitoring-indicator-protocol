package registry_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestAPIClient(t *testing.T) {
	t.Run("it registers an indicator document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		var receivedDoc v1.IndicatorDocument

		mux := http.NewServeMux()
		mux.HandleFunc("/v1/register", func(writer http.ResponseWriter, request *http.Request) {
			if request.Method == http.MethodPost {
				posted, _ := ioutil.ReadAll(request.Body)
				_ = json.Unmarshal(posted, &receivedDoc)

				writer.WriteHeader(200)
			} else {
				t.Fail()
			}
		})

		server := http.Server{
			Addr: "localhost:8975",
			Handler: mux,
		}

		go server.ListenAndServe()

		defer server.Close()
		_ = go_test.WaitForTCPServer("localhost:8975", time.Second)

		c := registry.NewAPIClient("http://localhost:8975", http.DefaultClient)

		sentDoc := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{{
					Product: "amazing-component",
				}},
			},
		}

		bytesToSend, _ := json.Marshal(sentDoc)

		err := c.AddIndicatorDocument(bytesToSend)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(receivedDoc).To(Equal(sentDoc))


	})

	t.Run("errors when registering document returns bad status code", func(t *testing.T) {
		g := NewGomegaWithT(t)

		mux := http.NewServeMux()
		mux.HandleFunc("/v1/register", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(400)
		})

		server := http.Server{
			Addr: "localhost:8976",
			Handler: mux,
		}

		go server.ListenAndServe()

		defer server.Close()
		_ = go_test.WaitForTCPServer("localhost:8976", time.Second)

		c := registry.NewAPIClient("http://localhost:8976", http.DefaultClient)

		sentDoc := v1.IndicatorDocument{
			Spec: v1.IndicatorDocumentSpec{
				Indicators: []v1.IndicatorSpec{{
					Product: "amazing-component",
				}},
			},
		}

		bytesToSend, _ := json.Marshal(sentDoc)

		err := c.AddIndicatorDocument(bytesToSend)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError("received non-successful response from registry: 400"))
	})

	t.Run("it parses the indicator response into Document Structs", func(t *testing.T) {
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

		c := registry.NewAPIClient("http://localhost:8975", http.DefaultClient)

		documents, e := c.IndicatorDocuments()
		g.Expect(e).ToNot(HaveOccurred())
		g.Expect(documents[0].Spec.Product.Name).To(Equal("my-component"))
		g.Expect(documents[0].Spec.Product.Version).To(Equal("1.2.3"))
	})
}
