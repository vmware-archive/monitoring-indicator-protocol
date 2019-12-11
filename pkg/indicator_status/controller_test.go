package indicator_status_test

import (
	"bytes"
	"log"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestStatusController(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	t.Run("updates all indicator statuses", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string][]float64{
			"rate(errors[5m])":  {50},
			"rate(happies[1m])": {9},
		})

		fakeRegistryClient := setupFakeRegistryClient([]registry.APIIndicatorResponse{
			{
				Name:   "error_rate",
				PromQL: "rate(errors[5m])",
				Thresholds: []registry.APIThresholdResponse{
					{
						Level:    "critical",
						Operator: "gte",
						Value:    50,
					},
				},
			}, {
				Name:   "happiness_rate",
				PromQL: "rate(happies[1m])",
				Thresholds: []registry.APIThresholdResponse{
					{
						Level:    "warning",
						Operator: "lt",
						Value:    10,
					},
				},
			},
		})

		controller := indicator_status.NewStatusController(
			fakeRegistryClient,
			fakeRegistryClient,
			fakeQueryClient,
			time.Minute,
		)

		go controller.Start()

		g.Eventually(fakeRegistryClient.countBulkUpdates).Should(Equal(1))

		g.Expect(fakeRegistryClient.statusesForUID("uaa-abc-123")).To(ConsistOf(
			registry.ApiV1UpdateIndicatorStatus{
				Name:   "error_rate",
				Status: test_fixtures.StrPtr("critical"),
			},
			registry.ApiV1UpdateIndicatorStatus{
				Name:   "happiness_rate",
				Status: test_fixtures.StrPtr("warning"),
			},
		))
	})

	t.Run("handles multiple series", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string][]float64{
			"rate(errors[5m])": {40, 51},
		})
		fakeRegistryClient := setupFakeRegistryClient([]registry.APIIndicatorResponse{
			{
				Name:   "error_rate",
				PromQL: "rate(errors[5m])",
				Thresholds: []registry.APIThresholdResponse{
					{
						Level:    "critical",
						Operator: "gte",
						Value:    50,
					},
					{
						Level:    "warning",
						Operator: "gte",
						Value:    30,
					},
				},
			},
		})

		controller := indicator_status.NewStatusController(
			fakeRegistryClient,
			fakeRegistryClient,
			fakeQueryClient,
			time.Minute,
		)

		go controller.Start()

		g.Eventually(fakeRegistryClient.countBulkUpdates).Should(Equal(1))

		g.Expect(fakeRegistryClient.statusesForUID("uaa-abc-123")).To(ConsistOf(
			registry.ApiV1UpdateIndicatorStatus{
				Name:   "error_rate",
				Status: test_fixtures.StrPtr("critical"),
			},
		))
	})

	t.Run("only queries indicators with thresholds", func(t *testing.T) {

		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string][]float64{
			"rate(errors[5m])":  {50},
			"rate(happies[1m])": {9},
		})

		fakeRegistryClient := setupFakeRegistryClient([]registry.APIIndicatorResponse{
			{
				Name:       "happiness_rate",
				PromQL:     "rate(happies[1m])",
				Thresholds: []registry.APIThresholdResponse{},
			},
		})

		controller := indicator_status.NewStatusController(
			fakeRegistryClient,
			fakeRegistryClient,
			fakeQueryClient,
			time.Minute,
		)

		go controller.Start()

		g.Consistently(fakeQueryClient.GetQueries).Should(HaveLen(0))
	})

}

func setupFakeQueryClientWithVectorResponses(responses map[string][]float64) *fakeQueryClient {
	fakeQueryClient := &fakeQueryClient{
		responses: responses,
	}
	return fakeQueryClient
}

func setupFakeRegistryClient(indicators []registry.APIIndicatorResponse) *fakeRegistryClient {
	var fakeRegistryClient = &fakeRegistryClient{
		receivedStatuses: map[string][]registry.ApiV1UpdateIndicatorStatus{},
		bulkUpdates:      0,
		indicatorDocuments: []registry.APIDocumentResponse{
			{
				UID:  "uaa-abc-123",
				Spec: registry.APIDocumentSpecResponse{Indicators: indicators},
			},
		},
	}
	return fakeRegistryClient
}

type fakeQueryClient struct {
	responses map[string][]float64
	queries   []string
	sync.Mutex
}

func (s *fakeQueryClient) QueryVectorValues(query string) ([]float64, error) {
	s.Lock()
	defer s.Unlock()
	s.queries = append(s.queries, query)

	return s.responses[query], nil
}

func (s *fakeQueryClient) GetQueries() []string {
	s.Lock()
	defer s.Unlock()
	return s.queries
}

type fakeRegistryClient struct {
	indicatorDocuments []registry.APIDocumentResponse

	sync.Mutex
	bulkUpdates      int
	receivedStatuses map[string][]registry.ApiV1UpdateIndicatorStatus
}

func (f *fakeRegistryClient) IndicatorDocuments() ([]registry.APIDocumentResponse, error) {
	return f.indicatorDocuments, nil
}

func (f *fakeRegistryClient) BulkStatusUpdate(statusUpdates []registry.ApiV1UpdateIndicatorStatus, documentId string) error {
	f.Lock()
	defer f.Unlock()

	f.bulkUpdates = f.bulkUpdates + 1
	f.receivedStatuses[documentId] = statusUpdates
	return nil
}

func (f *fakeRegistryClient) statusesForUID(uid string) []registry.ApiV1UpdateIndicatorStatus {
	f.Lock()
	defer f.Unlock()

	return f.receivedStatuses[uid]
}

func (f *fakeRegistryClient) countBulkUpdates() int {
	f.Lock()
	defer f.Unlock()

	return f.bulkUpdates
}
