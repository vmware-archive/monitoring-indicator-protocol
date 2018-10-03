package verification_test

import (
	. "github.com/onsi/gomega"
	"testing"

	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/verification"
)

func TestVerifyMetric(t *testing.T) {

	t.Run("returns a result", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return logCachePromQLResponse(3, 4), nil
			},
		}

		m := indicator.Indicator{
			Name:          "latency",
			PromQL:        `latency{source_id="demo_component",deployment="cf"}[1m]`,
		}

		_, err := verification.VerifyIndicator(m, client)

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("returns an error when the request fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return nil, fmt.Errorf("oh no! can't get a port!")
			},
		}

		m := indicator.Indicator{
			Name:          "latency",
			PromQL:        `latency{source_id="demo_component",deployment="cf"}[1m]`,
		}

		_, err := verification.VerifyIndicator(m, client)

		g.Expect(err).To(HaveOccurred())
	})
}

func logCachePromQLResponse(numSeries, numPoints int) model.Value {
	var series *model.SampleStream
	var seriesList model.Matrix
	for i := 0; i < numSeries; i++ {
		series = &model.SampleStream{
			Metric: model.Metric{
				"vm": model.LabelValue(fmt.Sprintf("vm-%d", i)),
			},
			Values: nil,
		}

		series.Values = make([]model.SamplePair, numPoints)
		for j := 0; j < numPoints; j++ {
			series.Values[j] = model.SamplePair{
				Value:     model.SampleValue(float64(j * i)),
				Timestamp: model.Time(time.Now().Unix()),
			}
		}

		seriesList = append(seriesList, series)
	}

	return seriesList
}

type mockQueryClient struct {
	TestQuerier func(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

func (m mockQueryClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	return m.TestQuerier(ctx, query, ts)
}
