package verification_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/prometheus/common/model"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/verification"
)

func TestVerifyMetric(t *testing.T) {

	t.Run("returns matrix results", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return matrixResponse(3, 4), nil
			},
		}

		m := v1.IndicatorSpec{
			Name:   "latency",
			PromQL: `latency{source_id="demo_component",deployment="cf"}[1m]`,
		}

		result, err := verification.VerifyIndicator(m, &client)

		g.Expect(result).To(Equal(verification.Result{
			MaxNumberOfPoints: 4,
			Series: []verification.ResultSeries{
				{
					Labels: "{vm=\"vm-4\"}",
					Points: []string{"0", "1", "4", "9"},
				},
				{
					Labels: "{vm=\"vm-4\"}",
					Points: []string{"0", "1", "4", "9"},
				},
				{
					Labels: "{vm=\"vm-4\"}",
					Points: []string{"0", "1", "4", "9"},
				},
			},
		}))

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("interpolates $step", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return matrixResponse(3, 4), nil
			},
		}

		m := v1.IndicatorSpec{
			Name:   "latency",
			PromQL: `latency{source_id="demo_component",deployment="cf"}[$step]`,
		}

		_, err := verification.VerifyIndicator(m, &client)

		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(client.Queries).To(HaveLen(1))
		g.Expect(client.Queries).To(ContainElement(`latency{source_id="demo_component",deployment="cf"}[1m]`))
	})

	t.Run("returns vector results", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return vectorResponse(3), nil
			},
		}

		m := v1.IndicatorSpec{
			Name:   "latency",
			PromQL: `latency{source_id="demo_component",deployment="cf"}[1m]`,
		}

		result, err := verification.VerifyIndicator(m, &client)

		g.Expect(result).To(Equal(verification.Result{
			MaxNumberOfPoints: 3,
			Series: []verification.ResultSeries{
				{
					Labels: "{vm=\"vm-3\"} => 0 @[0.1]",
					Points: []string{"0", "1", "2"},
				},
			},
		}))

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("returns an error when the request fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return nil, fmt.Errorf("oh no! can't get a port!")
			},
		}

		m := v1.IndicatorSpec{
			Name:   "latency",
			PromQL: `latency{source_id="demo_component",deployment="cf"}[1m]`,
		}

		_, err := verification.VerifyIndicator(m, &client)

		g.Expect(err).To(HaveOccurred())
	})
}

func vectorResponse(numPoints int) model.Vector {
	var vector model.Vector

	for i := 0; i < numPoints; i++ {
		sample := &model.Sample{
			Metric: model.Metric{
				"vm": model.LabelValue(fmt.Sprintf("vm-%d", numPoints)),
			},
			Value:     model.SampleValue(i),
			Timestamp: 100,
		}

		vector = append(vector, sample)
	}

	return vector
}

func matrixResponse(numSeries, numPoints int) model.Matrix {
	var seriesList model.Matrix
	for i := 0; i < numSeries; i++ {
		var series *model.SampleStream

		series = &model.SampleStream{
			Metric: model.Metric{
				"vm": model.LabelValue(fmt.Sprintf("vm-%d", numPoints)),
			},
			Values: nil,
		}

		series.Values = make([]model.SamplePair, numPoints)
		for j := 0; j < numPoints; j++ {
			series.Values[j] = model.SamplePair{
				Value:     model.SampleValue(float64(j * j)),
				Timestamp: model.Time(time.Now().Unix()),
			}
		}

		seriesList = append(seriesList, series)
	}

	return seriesList
}

type mockQueryClient struct {
	Queries     []string
	TestQuerier func(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

func (m *mockQueryClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	m.Queries = append(m.Queries, query)
	return m.TestQuerier(ctx, query, ts)
}
