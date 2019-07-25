package registry_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestRegisterHandler(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("valid requests return 200", func(t *testing.T) {
		t.Run("for apiVersion v0", func(t *testing.T) {
			body := bytes.NewBuffer([]byte(`
---
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
    gte: 50
`))

			req := httptest.NewRequest("POST", "/register", body)
			resp := httptest.NewRecorder()

			docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
			handle := registry.NewRegisterHandler(docStore)
			handle(resp, req)

			g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))

			g.Expect(resp.Code).To(Equal(http.StatusOK))
			g.Expect(docStore.AllDocuments()).To(ConsistOf(v1alpha1.IndicatorDocument{
				TypeMeta: v1.TypeMeta{
					Kind:       "IndicatorDocument",
					APIVersion: api_versions.V0,
				},
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"deployment": "redis-abc-123",
					},
				},

				Spec: v1alpha1.IndicatorDocumentSpec{
					Product: v1alpha1.Product{Name: "redis-tile", Version: "0.11"},
					Layout: v1alpha1.Layout{
						Sections: []v1alpha1.Section{{
							Title:      "Metrics",
							Indicators: []string{"test_performance_indicator"},
						}},
					},
					Indicators: []v1alpha1.IndicatorSpec{{
						Name:   "test_performance_indicator",
						PromQL: "prom",
						Alert: v1alpha1.Alert{
							For:  "1m",
							Step: "1m",
						},
						Presentation: v1alpha1.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
						Thresholds: []v1alpha1.Threshold{
							{
								Level:    "warning",
								Operator: v1alpha1.GreaterThanOrEqualTo,
								Value:    50,
							},
						},
					},
					}},
			}))
		})
		t.Run("for apiVersion v1alpha1", func(t *testing.T) {
			body := bytes.NewBuffer([]byte(`---
apiVersion: apps.pivotal.io/v1alpha1

kind: IndicatorDocument


metadata:
  labels:
    deployment: redis-abc-123

spec:
  product: 
    name: redis-tile
    version: 0.11
  
  indicators:
  - name: test_performance_indicator
    promql: prom
    thresholds:
    - level: warning
      operator: gte
      value: 50`))

			req := httptest.NewRequest("POST", "/register", body)
			resp := httptest.NewRecorder()

			docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
			handle := registry.NewRegisterHandler(docStore)
			handle(resp, req)

			g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
			g.Expect(resp.Code).To(Equal(http.StatusOK))
			g.Expect(docStore.AllDocuments()).To(ConsistOf(v1alpha1.IndicatorDocument{
				TypeMeta: v1.TypeMeta{
					Kind:       "IndicatorDocument",
					APIVersion: api_versions.V1alpha1,
				},
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"deployment": "redis-abc-123",
					},
				},
				Spec: v1alpha1.IndicatorDocumentSpec{
					Product: v1alpha1.Product{Name: "redis-tile", Version: "0.11"},
					Layout: v1alpha1.Layout{
						Sections: []v1alpha1.Section{{
							Title:      "Metrics",
							Indicators: []string{"test_performance_indicator"},
						}},
					},
					Indicators: []v1alpha1.IndicatorSpec{{
						Name:   "test_performance_indicator",
						PromQL: "prom",
						Alert: v1alpha1.Alert{
							For:  "1m",
							Step: "1m",
						},
						Presentation: v1alpha1.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
						Thresholds: []v1alpha1.Threshold{
							{
								Level:    "warning",
								Operator: v1alpha1.GreaterThanOrEqualTo,
								Value:    50,
							},
						},
					},
					}},
			}))
		})
	})

	t.Run("it returns 400 if there are validation errors", func(t *testing.T) {
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
		body := bytes.NewBuffer([]byte(`---
apiVersion: {
}`))

		req := httptest.NewRequest("POST", "/register?deployment=redis-abc&product=redis-tile", body)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		handle := registry.NewRegisterHandler(docStore)
		handle(resp, req)

		g.Expect(docStore.AllDocuments()).To(HaveLen(0))

		g.Expect(resp.Code).To(Equal(http.StatusBadRequest))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBody).To(MatchJSON(`{ "errors": ["could not unmarshal apiVersion"] }`))
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
		docStore.UpsertDocument(v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: api_versions.V1alpha1,
				Kind:       "IndicatorDocument",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"deployment": "abc-123",
				},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "my-product-a", Version: "1"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "indie1",
					PromQL: "promql1",
					Alert: v1alpha1.Alert{
						For:  "5m",
						Step: "10s",
					},
					ServiceLevel: nil,
					Presentation: test_fixtures.DefaultPresentation(),
				}, {
					Name:   "indie2",
					PromQL: "promql2",
					Alert: v1alpha1.Alert{
						For:  "5m",
						Step: "10s",
					},
					ServiceLevel: &v1alpha1.ServiceLevel{
						Objective: float64(99.99),
					},
					Presentation: v1alpha1.Presentation{
						ChartType:    "status",
						CurrentValue: false,
						Frequency:    0,
						Units:        "nanoseconds",
					},
				}},
			},
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

		expectedJSON, err := ioutil.ReadFile("test_fixtures/example_response2.json")
		g.Expect(responseBody).To(MatchJSON(expectedJSON))
	})

	t.Run("allows various characters in document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		req := httptest.NewRequest("POST", "/indicator-documents", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		docStore.UpsertDocument(v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: api_versions.V1alpha1,
				Kind:       "IndicatorDocument",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"deployment % ": "%d abc-123",
				},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{
				Product: v1alpha1.Product{Name: "my-product-a", Version: "1"},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:          "indie1",
					PromQL:        "promql1",
					Alert:         test_fixtures.DefaultAlert(),
					Documentation: map[string]string{" % ": "{0}", "%n": "%*.*s"},
					Presentation:  test_fixtures.DefaultPresentation(),
				}},
			},
		})

		statusStore := status_store.New(func() time.Time { return time.Date(2012, 12, 1, 16, 45, 19, 0, time.UTC) })

		handle := registry.NewIndicatorDocumentsHandler(docStore, statusStore)
		handle(resp, req)

		g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		g.Expect(resp.Code).To(Equal(http.StatusOK))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		expectedJSON, err := ioutil.ReadFile("test_fixtures/example_response3.json")
		g.Expect(responseBody).To(MatchJSON(expectedJSON))

	})
}
