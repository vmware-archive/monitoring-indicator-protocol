package prometheus_client_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_client"

	"github.com/prometheus/common/model"
)

func TestPrometheusClient(t *testing.T) {
	t.Run("when Prometheus returns an empty vector, returns empty values", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := spyPromqlClient{
			response: model.Vector{},
		}
		client := prometheus_client.PrometheusClient{
			Api: &fakeQueryClient,
		}

		values, err := client.QueryVectorValues("rate(errors[5m])")

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(values).To(BeEmpty())
	})

	t.Run("when Prometheus returns a non-empty vector, returns non-empty values", func(t *testing.T) {
		g := NewGomegaWithT(t)

		expectedResponse := []float64{9, 10, 11, 12}

		fakeQueryClient := spyPromqlClient{
			response: vectorResponse(expectedResponse),
		}
		client := prometheus_client.PrometheusClient{
			Api: &fakeQueryClient,
		}

		values, err := client.QueryVectorValues("rate(errors[5m])")

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(values).To(Equal(expectedResponse))
	})

	t.Run("when Prometheus returns an error, returns the error", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := spyPromqlClient{
			error: errors.New("test-error"),
		}
		client := prometheus_client.PrometheusClient{
			Api: &fakeQueryClient,
		}

		values, err := client.QueryVectorValues("rate(errors[5m])")

		g.Expect(err).To(HaveOccurred())
		g.Expect(values).To(BeNil())
	})

	t.Run("when Prometheus returns a non-vector value, returns the error", func(t *testing.T) {
		g := NewGomegaWithT(t)

		expectedResponse := &model.Scalar{}

		fakeQueryClient := spyPromqlClient{
			response: expectedResponse,
		}
		client := prometheus_client.PrometheusClient{
			Api: &fakeQueryClient,
		}
		values, err := client.QueryVectorValues("rate(errors[5m])")

		g.Expect(values).To(BeNil())
		g.Expect(err.Error()).To(Equal("could not assert result from prometheus as Vector: *model.Scalar"))
	})

	t.Run("it queries prometheus with proper context", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := spyPromqlClient{
			response: model.Vector{},
		}
		client := prometheus_client.PrometheusClient{
			Api: &fakeQueryClient,
		}

		_, err := client.QueryVectorValues("rate(errors[5m])")
		g.Expect(err).NotTo(HaveOccurred())

		ctx := fakeQueryClient.ctxs[0]
		g.Expect(ctx.Done()).To(BeClosed())
		_, hasDeadline := ctx.Deadline()
		g.Expect(hasDeadline).To(BeTrue())
	})
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

//********** Spy Prometheus Client **********//

type spyPromqlClient struct {
	response model.Value
	queries  []string
	ctxs     []context.Context
	error    error
	sync.Mutex
}

func (s *spyPromqlClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	s.Lock()
	defer s.Unlock()
	if s.error != nil {
		return nil, s.error
	}
	s.queries = append(s.queries, query)
	s.ctxs = append(s.ctxs, ctx)

	return s.response, nil
}
