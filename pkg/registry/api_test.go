package registry_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benjamintf1/unmarshalledmatchers"
	"github.com/gorilla/mux"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestRegisterHandler(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("valid requests return 200", func(t *testing.T) {
		t.Run("for apiVersion v1", func(t *testing.T) {
			body := bytes.NewBuffer([]byte(`---
apiVersion: indicatorprotocol.io/v1

kind: IndicatorDocument


metadata:
  labels:
    deployment: redis-abc-123

spec:
  product: 
    name: redis-tile
    version: v0.11
  
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
			g.Expect(docStore.AllDocuments()).To(ConsistOf(v1.IndicatorDocument{
				TypeMeta: metaV1.TypeMeta{
					Kind:       "IndicatorDocument",
					APIVersion: api_versions.V1,
				},
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						"deployment": "redis-abc-123",
					},
				},
				Spec: v1.IndicatorDocumentSpec{
					Product: v1.Product{Name: "redis-tile", Version: "v0.11"},
					Layout: v1.Layout{
						Title: "redis-tile - v0.11",
						Sections: []v1.Section{{
							Title:      "Metrics",
							Indicators: []string{"test_performance_indicator"},
						}},
					},
					Indicators: []v1.IndicatorSpec{{
						Name:   "test_performance_indicator",
						PromQL: "prom",
						Presentation: v1.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
						Thresholds: []v1.Threshold{
							{
								Level:    "warning",
								Operator: v1.GreaterThanOrEqualTo,
								Value:    50,
								Alert: v1.Alert{
									For:  "1m",
									Step: "1m",
								},
							},
						},
					},
					}},
			}))
		})
	})

	t.Run("it returns 400 if there are validation errors", func(t *testing.T) {
		body := bytes.NewBuffer([]byte(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  indicators:
  - promql: ""
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

		g.Expect(responseBody).To(unmarshalledmatchers.MatchUnorderedJSON(`{ "errors": [
	"IndicatorDocument.spec.product.name in body should be at least 1 chars long", 
	"IndicatorDocument.spec.product.version in body should be at least 1 chars long", 
	"IndicatorDocument.spec.indicators.promql in body should be at least 1 chars long",
    "indicators[0] is invalid by schema: IndicatorSpec.promql in body should be at least 1 chars long"
]}`))
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

		type errResponse struct {
			Errors []string `json:"errors"`
		}

		var errResp errResponse
		err = json.Unmarshal(responseBody, &errResp)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(errResp.Errors).To(ConsistOf("failed to parse metadata, could not read apiVersion: could not unmarshal apiVersion, check that document contains valid YAML"))
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

		req := httptest.NewRequest("GET", "/indicator-documents", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		docStore.UpsertDocument(makeIndicatorDocument(map[string]string{
			"deployment": "abc-123",
		}))

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

	t.Run("it allows filtering by metadata keys", func(t *testing.T) {
		g := NewGomegaWithT(t)

		req := httptest.NewRequest("GET", "/indicator-documents?source_id=test-id", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		docStore.UpsertDocument(makeIndicatorDocument(map[string]string{
			"deployment": "abc-12345",
			"source_id":  "not-test-id",
		}))
		docStore.UpsertDocument(makeIndicatorDocument(map[string]string{
			"deployment": "abc-123",
			"source_id":  "test-id",
		}))

		statusStore := status_store.New(func() time.Time { return time.Date(2012, 12, 1, 16, 45, 19, 0, time.UTC) })
		statusStore.UpdateStatus(status_store.UpdateRequest{
			Status:        test_fixtures.StrPtr("critical"),
			IndicatorName: "indie2",
			DocumentUID:   "my-product-a-82a34f2cefd4899aee804e19e39aab95c0420ec3",
		})

		handle := registry.NewIndicatorDocumentsHandler(docStore, statusStore)
		handle(resp, req)

		g.Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		g.Expect(resp.Code).To(Equal(http.StatusOK))

		responseBody, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		expectedJSON, err := ioutil.ReadFile("test_fixtures/example_response5.json")
		g.Expect(responseBody).To(MatchJSON(expectedJSON))
	})

	t.Run("returns empty list when cannot match on query parameter", func(t *testing.T) {
		g := NewGomegaWithT(t)

		req := httptest.NewRequest("GET", "/indicator-documents?source_id=test-id", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		docStore.UpsertDocument(makeIndicatorDocument(map[string]string{
			"deployment": "abc-123",
		}))

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

		g.Expect(responseBody).To(MatchJSON("[]"))
	})

	t.Run("allows various characters in document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		req := httptest.NewRequest("POST", "/indicator-documents", nil)
		resp := httptest.NewRecorder()

		docStore := registry.NewDocumentStore(1*time.Minute, time.Now)
		docStore.UpsertDocument(v1.IndicatorDocument{
			TypeMeta: metaV1.TypeMeta{
				APIVersion: api_versions.V1,
				Kind:       "IndicatorDocument",
			},
			ObjectMeta: metaV1.ObjectMeta{
				Labels: map[string]string{
					"deployment % ": "%d abc-123",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{Name: "my-product-a", Version: "1"},
				Indicators: []v1.IndicatorSpec{{
					Name:          "indie1",
					PromQL:        "promql1",
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

func makeIndicatorDocument(labels map[string]string) v1.IndicatorDocument {
	return v1.IndicatorDocument{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: api_versions.V1,
			Kind:       "IndicatorDocument",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Labels: labels,
		},
		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{Name: "my-product-a", Version: "1"},
			Indicators: []v1.IndicatorSpec{{
				Name:   "indie1",
				PromQL: "promql1",
				Thresholds: []v1.Threshold{{
					Level:    "uh-oh",
					Operator: v1.EqualTo,
					Value:    1000,
					Alert:    v1.Alert{
						For:  "100h",
						Step: "9m",
					},
				}},
				Presentation: test_fixtures.DefaultPresentation(),
			}, {
				Name:   "indie2",
				PromQL: "promql2",
				Presentation: v1.Presentation{
					ChartType:    "status",
					CurrentValue: false,
					Frequency:    0,
					Units:        "nanoseconds",
				},
			}},
		},
	}
}
