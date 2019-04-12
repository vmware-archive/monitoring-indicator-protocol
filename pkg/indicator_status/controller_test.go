package indicator_status_test

import (
	"bytes"
	"context"
	"log"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	"github.com/prometheus/common/model"
)

func TestStatusController(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	t.Run("updates all indicator statuses", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])":  vectorResponse([]float64{50}),
			"rate(happies[1m])": vectorResponse([]float64{9}),
		})

		fakeRegistryClient := setupFakeRegistryClient([]registry.APIV0Indicator{
			{
				Name:   "error_rate",
				PromQL: "rate(errors[5m])",
				Thresholds: []registry.APIV0Threshold{
					{
						Level:    "critical",
						Operator: "gte",
						Value:    50,
					},
				},
			}, {
				Name:   "happiness_rate",
				PromQL: "rate(happies[1m])",
				Thresholds: []registry.APIV0Threshold{
					{
						Level:    "warning",
						Operator: "lt",
						Value:    10,
					},
				},
			},
		})

		controller := indicator_status.StatusController{
			RegistryClient: fakeRegistryClient,
			IntervalTime:   time.Minute,
			PromQLClient:   fakeQueryClient,
		}

		go controller.Start()

		g.Eventually(fakeRegistryClient.countBulkUpdates).Should(Equal(1))

		g.Expect(fakeRegistryClient.statusesForUID("uaa-abc-123")).To(ConsistOf(
			registry.APIV0UpdateIndicatorStatus{
				Name:   "error_rate",
				Status: test_fixtures.StrPtr("critical"),
			},
			registry.APIV0UpdateIndicatorStatus{
				Name:   "happiness_rate",
				Status: test_fixtures.StrPtr("warning"),
			},
		))
	})

	t.Run("handles multiple series", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])": vectorResponse([]float64{40, 51}),
		})
		fakeRegistryClient := setupFakeRegistryClient([]registry.APIV0Indicator{
			{
				Name:   "error_rate",
				PromQL: "rate(errors[5m])",
				Thresholds: []registry.APIV0Threshold{
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

		controller := indicator_status.StatusController{
			RegistryClient: fakeRegistryClient,
			IntervalTime:   time.Minute,
			PromQLClient:   fakeQueryClient,
		}

		go controller.Start()

		g.Eventually(fakeRegistryClient.countBulkUpdates).Should(Equal(1))

		g.Expect(fakeRegistryClient.statusesForUID("uaa-abc-123")).To(ConsistOf(
			registry.APIV0UpdateIndicatorStatus{
				Name:   "error_rate",
				Status: test_fixtures.StrPtr("critical"),
			},
		))
	})

	t.Run("ignores query results that are not vectors", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])":  matrixResponse(),
			"rate(happies[1m])": vectorResponse([]float64{9}),
		})

		fakeRegistryClient := setupFakeRegistryClient([]registry.APIV0Indicator{
			{
				Name:   "error_rate",
				PromQL: "rate(errors[5m])",
				Thresholds: []registry.APIV0Threshold{
					{
						Level:    "critical",
						Operator: "gte",
						Value:    50,
					},
				},
			}, {
				Name:   "happiness_rate",
				PromQL: "rate(happies[1m])",
				Thresholds: []registry.APIV0Threshold{
					{
						Level:    "warning",
						Operator: "lt",
						Value:    10,
					},
				},
			},
		})

		controller := indicator_status.StatusController{
			RegistryClient: fakeRegistryClient,
			IntervalTime:   time.Minute,
			PromQLClient:   fakeQueryClient,
		}

		go controller.Start()

		g.Eventually(fakeRegistryClient.countBulkUpdates).Should(Equal(1))

		g.Expect(fakeRegistryClient.statusesForUID("uaa-abc-123")).To(ConsistOf(
			registry.APIV0UpdateIndicatorStatus{
				Name:   "happiness_rate",
				Status: test_fixtures.StrPtr("warning"),
			},
		))
	})

	t.Run("only queries indicators with thresholds", func(t *testing.T) {

		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])":  vectorResponse([]float64{50}),
			"rate(happies[1m])": vectorResponse([]float64{9}),
		})

		fakeRegistryClient := setupFakeRegistryClient([]registry.APIV0Indicator{
			{
				Name:       "happiness_rate",
				PromQL:     "rate(happies[1m])",
				Thresholds: []registry.APIV0Threshold{},
			},
		})

		controller := indicator_status.StatusController{
			RegistryClient: fakeRegistryClient,
			IntervalTime:   time.Minute,
			PromQLClient:   fakeQueryClient,
		}

		go controller.Start()

		g.Consistently(fakeQueryClient.QueryCount).Should(Equal(0))
	})

}

func setupFakeQueryClientWithVectorResponses(responses map[string]model.Value) *fakeQueryClient {
	fakeQueryClient := &fakeQueryClient{
		responses: responses,
	}
	return fakeQueryClient
}

func setupFakeRegistryClient(indicators []registry.APIV0Indicator) *fakeRegistryClient {
	var fakeRegistryClient = &fakeRegistryClient{
		receivedStatuses: map[string][]registry.APIV0UpdateIndicatorStatus{},
		bulkUpdates:      0,
		indicatorDocuments: []registry.APIV0Document{
			{
				UID:        "uaa-abc-123",
				Indicators: indicators,
			},
		},
	}
	return fakeRegistryClient
}

func matrixResponse() model.Matrix {
	var seriesList model.Matrix
	var series *model.SampleStream

	series = &model.SampleStream{
		Metric: model.Metric{},
		Values: nil,
	}

	series.Values = []model.SamplePair{{
		Value:     model.SampleValue(float64(100)),
		Timestamp: model.Time(time.Now().Unix()),
	}}

	seriesList = append(seriesList, series)

	return seriesList
}

func vectorResponse(values []float64) model.Vector {
	var vector model.Vector

	for _, v := range values {
		sample := &model.Sample{
			Metric: model.Metric{
				"deployment": "uaa123",
			},
			Value:     model.SampleValue(v),
			Timestamp: 100,
		}

		vector = append(vector, sample)
	}
	return vector
}

type fakeQueryClient struct {
	sync.Mutex
	responses  map[string]model.Value
	queryCount int
}

func (m *fakeQueryClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	m.Lock()
	defer m.Unlock()

	m.queryCount += 1
	return m.responses[query], nil
}

func (m *fakeQueryClient) QueryCount() int {
	m.Lock()
	defer m.Unlock()

	return m.queryCount
}

type fakeRegistryClient struct {
	indicatorDocuments []registry.APIV0Document

	sync.Mutex
	bulkUpdates      int
	receivedStatuses map[string][]registry.APIV0UpdateIndicatorStatus
}

func (f *fakeRegistryClient) IndicatorDocuments() ([]registry.APIV0Document, error) {
	return f.indicatorDocuments, nil
}

func (f *fakeRegistryClient) BulkStatusUpdate(statusUpdates []registry.APIV0UpdateIndicatorStatus, documentId string) error {
	f.Lock()
	defer f.Unlock()

	f.bulkUpdates = f.bulkUpdates + 1
	f.receivedStatuses[documentId] = statusUpdates
	return nil
}

func (f *fakeRegistryClient) statusesForUID(uid string) []registry.APIV0UpdateIndicatorStatus {
	f.Lock()
	defer f.Unlock()

	return f.receivedStatuses[uid]
}

func (f *fakeRegistryClient) countBulkUpdates() int {
	f.Lock()
	defer f.Unlock()

	return f.bulkUpdates
}
