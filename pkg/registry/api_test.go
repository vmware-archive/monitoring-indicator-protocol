package registry_test

import (
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"testing"

	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/registry"
)

func TestRegisterHandler(t *testing.T) {
	t.Run("it returns 200 if the request is valid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
apiVersion: v0

product: 
  name: redis-tile
  version: 0.11

metadata:
  deployment: redis-abc-123

indicators:
- name: test_performance_indicator
  promql: prom
  thresholds:
  - level: warning
    gte: 50`))

		req := httptest.NewRequest("POST", "/register", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1 * time.Minute)
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		g.Expect(resp.Code).To(Equal(http.StatusOK))
		g.Expect(docStore.All()).To(ConsistOf(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "redis-tile", Version: "0.11"},
			Metadata: map[string]string{
				"deployment": "redis-abc-123",
			},
			Indicators: []indicator.Indicator{{
				Name:   "test_performance_indicator",
				PromQL: "prom",
				Thresholds: []indicator.Threshold{
					{
						Level:    "warning",
						Operator: indicator.GreaterThanOrEqualTo,
						Value:    50,
					},
				},
			}},
		}))
	})

	t.Run("it returns 422 if there are validation errors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
apiVersion: v0
indicators:
- promql: " "
  name: none
`))

		req := httptest.NewRequest("POST", "/register?deployment=redis-abc", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1 * time.Minute)
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.All()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusUnprocessableEntity))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`{ "errors": ["product name is required", "product version is required", "indicators[0] promql is required"]}`))
	})

	t.Run("it returns 400 if the yml is invalid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
indicators: aasdfasdf`))

		req := httptest.NewRequest("POST", "/register?deployment=redis-abc&product=redis-tile", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1 * time.Minute)
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.All()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusBadRequest))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`{ "errors": ["could not unmarshal indicators: yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str ` + "`aasdfasdf`" + ` into []indicator.yamlIndicator"] }`))
	})
}

func TestIndicatorDocumentsHandler(t *testing.T) {
	t.Run("it returns 200", func(t *testing.T) {
		g := NewGomegaWithT(t)

		req := httptest.NewRequest("POST", "/indicator-documents", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1 * time.Minute)
		docStore.Upsert(indicator.Document{
			Product: indicator.Product{Name: "my-product-a", Version: "1"},
			Metadata: map[string]string{
				"deployment": "abc-123",
			},
			Indicators: []indicator.Indicator{{
				Name: "test_errors1",
			}, {
				Name:         "test_errors2",
			}},
		})

		handle := registry.NewIndicatorDocumentsHandler(docStore)
		handle(resp, req)

		g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		g.Expect(resp.Code).To(Equal(http.StatusOK))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`
			[
 				{
                    "apiVersion": "",
                    "product": {
						"name": "my-product-a",
                    	"version": "1"
					},
                    "metadata": {
                      "deployment": "abc-123"
                    },
                    "indicators": [
                      {
                        "name": "test_errors1",
                        "promql": "",
                        "thresholds": [],
                        "documentation": null
                      },
                      {
                        "name": "test_errors2",
                        "promql": "",
                        "thresholds": [],
                        "documentation": null
                      }
                    ],
                    "layout": {
                      "title": "",
                      "description": "",
                      "sections": [],
                      "owner": ""
                    }
                  }
			]`))
	})
}
