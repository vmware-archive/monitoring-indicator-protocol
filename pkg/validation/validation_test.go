package validation_test

import (
	. "github.com/onsi/gomega"
	"testing"

	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/validation"
)

func TestVerifyMetric(t *testing.T) {

	t.Run("returns a result", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return logCachePromQLResponse(3, 4), nil
			},
		}

		m := indicator.Metric{
			Title:       "Demo Component",
			Origin:      "demo",
			SourceID:    "demo_component",
			Name:        "latency",
			Type:        "gauge",
			Description: "A test metric",
		}

		_, err := validation.VerifyMetric(m, `latency{source_id="demo_component",deployment="cf"}[1m]`, client)

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("returns an error when the request fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{
			TestQuerier: func(ctx context.Context, query string, ts time.Time) (model.Value, error) {
				return nil, fmt.Errorf("oh no! can't get a port!")
			},
		}

		m := indicator.Metric{
			Title:       "Demo Component",
			Origin:      "demo",
			SourceID:    "demo_component",
			Name:        "latency",
			Type:        "gauge",
			Description: "A test metric",
		}

		_, err := validation.VerifyMetric(m, `latency{source_id="demo_component",deployment="cf"}[1m]`, client)

		g.Expect(err).To(HaveOccurred())
	})
}

func TestFormatQuery(t *testing.T) {

	var characterConversions = []struct {
		input       indicator.Metric
		expectation string
	}{
		{
			input:       indicator.Metric{SourceID: "router", Name: "uaa.latency"},
			expectation: `uaa_latency{source_id="router",deployment="cf"}[1m]`,
		},
		{
			input:       indicator.Metric{SourceID: "router", Name: `uaa/latency\a`},
			expectation: `uaa_latency_a{source_id="router",deployment="cf"}[1m]`,
		},
		{
			input:       indicator.Metric{SourceID: "router", Name: "uaa-latency"},
			expectation: `uaa_latency{source_id="router",deployment="cf"}[1m]`,
		},
	}

	for _, cc := range characterConversions {
		t.Run(cc.input.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(validation.FormatQuery(cc.input, "cf")).To(Equal(cc.expectation))
		})
	}
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
