package registry_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
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

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		g.Expect(resp.Code).To(Equal(http.StatusOK))
		g.Expect(docStore.AllDocuments()).To(ConsistOf(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "redis-tile", Version: "0.11"},
			Metadata: map[string]string{
				"deployment": "redis-abc-123",
			},
			Layout: indicator.Layout{
				Sections: []indicator.Section{{
					Title:      "Metrics",
					Indicators: []string{"test_performance_indicator"},
				}},
			},
			Indicators: []indicator.Indicator{{
				Name:   "test_performance_indicator",
				PromQL: "prom",
				Alert: indicator.Alert{
					For:  "1m",
					Step: "1m",
				},
				Presentation: indicator.Presentation{
					CurrentValue: false,
					ChartType:    "step",
					Frequency:    0,
					Labels:       []string{},
				},
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

	t.Run("it returns 400 if there are validation errors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`---
apiVersion: v0
indicators:
- promql: " "
  name: none
`))

		req := httptest.NewRequest("POST", "/register?deployment=redis-abc", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.AllDocuments()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusBadRequest))

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

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.AllDocuments()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusBadRequest))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`{ "errors": ["could not unmarshal indicators: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str ` + "`aasdfasdf`" + ` into []indicator.yamlIndicator"] }`))
	})
}

func TestBulkStatusUpdateHandler(t *testing.T) {
	t.Run("it returns 200 & updates the store status", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`[{"name": "latency", "status": "critical"},{"name": "error_rate", "status": "warning"}]`))
		req := httptest.NewRequest("POST", "/", body)
		resp := httptest.NewRecorder()

		now := time.Now()
		store := status_store.New(func() time.Time { return now })

		req = mux.SetURLVars(req, map[string]string{
			"documentID": "my-component-1234234234",
		})
		registry.NewIndicatorStatusBulkUpdateHandler(store)(resp, req)

		g.Expect(store.StatusFor("my-component-1234234234", "latency")).To(Equal(status_store.IndicatorStatus{
			DocumentUID:   "my-component-1234234234",
			IndicatorName: "latency",
			Status:        test_fixtures.StrPtr("critical"),
			UpdatedAt:     now,
		}))

		g.Expect(store.StatusFor("my-component-1234234234", "error_rate")).To(Equal(status_store.IndicatorStatus{
			DocumentUID:   "my-component-1234234234",
			IndicatorName: "error_rate",
			Status:        test_fixtures.StrPtr("warning"),
			UpdatedAt:     now,
		}))
	})

	t.Run("it returns a 400 if indicator status are not passed into the body", func(t *testing.T) {
		g := NewGomegaWithT(t)

		body := bytes.NewBuffer([]byte(`[""]`))
		req := httptest.NewRequest("POST", "/", body)
		resp := httptest.NewRecorder()

		now := time.Now()
		store := status_store.New(func() time.Time { return now })

		req = mux.SetURLVars(req, map[string]string{
			"documentID": "my-component-1234234234",
		})
		registry.NewIndicatorStatusBulkUpdateHandler(store)(resp, req)

		g.Expect(resp.Result().StatusCode).To(Equal(http.StatusBadRequest))
	})
}

func TestIndicatorDocumentsHandler(t *testing.T) {
	t.Run("it returns 200", func(t *testing.T) {
		g := NewGomegaWithT(t)

		req := httptest.NewRequest("POST", "/indicator-documents", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		docStore.UpsertDocument(indicator.Document{
			Product: indicator.Product{Name: "my-product-a", Version: "1"},
			Metadata: map[string]string{
				"deployment": "abc-123",
			},
			Indicators: []indicator.Indicator{{
				Name:   "indie1",
				PromQL: "promql1",
				Alert: indicator.Alert{
					For:  "5m",
					Step: "10s",
				},
				ServiceLevel: nil,
				Presentation: test_fixtures.DefaultPresentation(),
			}, {
				Name:   "indie2",
				PromQL: "promql2",
				Alert: indicator.Alert{
					For:  "5m",
					Step: "10s",
				},
				ServiceLevel: &indicator.ServiceLevel{
					Objective: float64(99.99),
				},
				Presentation: indicator.Presentation{
					ChartType:    "status",
					CurrentValue: false,
					Frequency:    0,
					Units:        "nanoseconds",
				},
			}},
		})

		statusStore := status_store.New(func() time.Time { return time.Date(2012, 12, 1, 16, 45, 19, 0, time.UTC) })
		statusStore.UpdateStatus(status_store.UpdateRequest{
			Status:        test_fixtures.StrPtr("critical"),
			IndicatorName: "indie2",
			DocumentUID:   "my-product-a-a902332065d69c1787f419e235a1f1843d98c884",
		})

		handle := registry.NewIndicatorDocumentsHandler(docStore, statusStore)
		handle(resp, req)

		g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		g.Expect(resp.Code).To(Equal(http.StatusOK))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`
			[
 				{
                    "apiVersion": "",
					"uid": "my-product-a-a902332065d69c1787f419e235a1f1843d98c884",
                    "product": {
						"name": "my-product-a",
                    	"version": "1"
					},
                    "metadata": {
                      "deployment": "abc-123"
                    },
                    "indicators": [
                      {
                        "name": "indie1",
                        "promql": "promql1",
                        "thresholds": [],
						"alert": {
							"for": "5m",
							"step": "10s"
						},
						"serviceLevel": null,
                        "presentation": {
                          "chartType": "step",
                          "currentValue": false,
                          "frequency": 0,
                          "labels": [],
                          "units": ""
                        },
                        "status": null
                      },
                      {
                        "name": "indie2",
                        "promql": "promql2",
                        "thresholds": [],
						"alert": {
							"for": "5m",
							"step": "10s"
						},
						"serviceLevel": {
							"objective": 99.99
						},
                        "presentation": {
                          "chartType": "status",
                          "currentValue": false,
                          "frequency": 0,
                          "labels": [],
                          "units": "nanoseconds"
                        },
						"status": {
						  "value": "critical",
						  "updatedAt": "2012-12-01T16:45:19Z"
						}
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

	t.Run("allows various characters in document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		req := httptest.NewRequest("POST", "/indicator-documents", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		docStore.UpsertDocument(indicator.Document{
			Product: indicator.Product{Name: "my-product-a", Version: "1"},
			Metadata: map[string]string{
				"deployment % ": "%d abc-123",
			},
			Indicators: []indicator.Indicator{{
				Name:          "indie1",
				PromQL:        "promql1",
				Alert:         test_fixtures.DefaultAlert(),
				Documentation: map[string]string{" % ": "{0}", "%n": "%*.*s"},
				Presentation:  test_fixtures.DefaultPresentation(),
			}},
		})

		statusStore := status_store.New(func() time.Time { return time.Date(2012, 12, 1, 16, 45, 19, 0, time.UTC) })

		handle := registry.NewIndicatorDocumentsHandler(docStore, statusStore)
		handle(resp, req)

		g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		g.Expect(resp.Code).To(Equal(http.StatusOK))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`
			[
 				{
                    "apiVersion": "",
					"uid": "my-product-a-b5cf9762d7a2e7cec0d6a1e0b959149595c3c198",
                    "product": {
						"name": "my-product-a",
                    	"version": "1"
					},
                    "metadata": {
                      "deployment % ": "%d abc-123"
                    },
                    "indicators": [
                      {
                        "name": "indie1",
                        "promql": "promql1",
                        "thresholds": [],
						"alert": {
							"for": "1m",
							"step": "1m"
						},
						"serviceLevel": null,
                        "presentation": {
                          "chartType": "step",
                          "currentValue": false,
                          "frequency": 0,
                          "labels": [],
                          "units": ""
                        },
						"documentation": {
						  " % ": "{0}",
						  "%n": "%*.*s"
						},
                        "status": null
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
