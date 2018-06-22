package registry_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cf-indicators/pkg/registry"
	"net/http/httptest"
	"bytes"
	"net/http"
	"io/ioutil"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
)

func TestRegisterHandler(t *testing.T) {
	t.Run("it returns 200 if the request is valid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
metrics:
- name: latency
  source_id: demo
  origin: demo
  title: Demo Latency
  type: metricType
  frequency: 60s
  description: A test metric for testing

indicators:
- name: test_performance_indicator
  title: Test Performance Indicator
  metrics:
  - name: latency
    source_id: demo
  measurement: Measurement Text
  promql: prom
  thresholds:
  - level: warning
    gte: 50
    dynamic: true
  description: This is a valid markdown description.
  response: Panic!`))

		req := httptest.NewRequest("POST", "/register?deployment=redis-abc&product=redis-tile", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore()
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(resp.Code).To(Equal(http.StatusOK))
		g.Expect(docStore.All()).To(HaveLen(1))
		g.Expect(docStore.All()[0].Labels).To(Equal(map[string]string{
			"deployment": "redis-abc",
			"product": "redis-tile",
		}))
		g.Expect(docStore.All()[0].Indicators[0].Name).To(Equal("test_performance_indicator"))
	})

	t.Run("it returns 400 if a label has multiple values", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
metrics: []`))

		req := httptest.NewRequest("POST", "/register?test-label=redis-abc&test-label=other-value", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore()
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.All()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusBadRequest))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`{ "errors": ["label test-label has too many values"] }`))
	})

	t.Run("it returns 422 if there are validation errors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
metrics:
- name: latency
  source_id: demo
  origin: demo
  title: Demo Latency
  type: metricType
  frequency: 60s
  description: " "`))

		req := httptest.NewRequest("POST", "/register?deployment=redis-abc&product=redis-tile", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore()
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.All()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusUnprocessableEntity))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`{ "errors": ["metrics[0] description is required"] }`))
	})

	t.Run("it returns 400 if the yml is invalid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
metrics: aasdfasdf
- name: latency
  source_id: demo
  origin: demo
  title: Demo Latency
  type: metricType
  frequency: 60s
  description: `))

		req := httptest.NewRequest("POST", "/register?deployment=redis-abc&product=redis-tile", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore()
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.All()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusBadRequest))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`{ "errors": ["could not unmarshal indicators: yaml: line 2: did not find expected key"] }`))
	})
}

func TestIndicatorDocumentsHandler(t *testing.T) {
	t.Run("it returns 200", func(t *testing.T) {
		g:= NewGomegaWithT(t)

		req := httptest.NewRequest("POST", "/indicator-documents", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore()
		docStore.Upsert(map[string]string{"test-label":"test-value"}, []indicator.Indicator{{
			Name:        "test indicator",
		}})

		handle := registry.NewIndicatorDocumentsHandler(docStore)
		handle(resp, req)

		g.Expect(resp.Code).To(Equal(http.StatusOK))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`
			[
				{
					"labels": {
						"test-label": "test-value"
					},
					"indicators": [
						{
    			            "name": "test indicator",
    			            "title": "",
    			            "description": "",
    			            "promql": "",
    			            "thresholds": [],
    			            "metrics": [],
    			            "response": "",
    			            "measurement": ""
						}
					]
				}
			]`))
	})
}