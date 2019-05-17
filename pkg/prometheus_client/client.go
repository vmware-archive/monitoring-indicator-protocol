package prometheus_client

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

type PrometheusQueryAPI interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

type PrometheusClient struct {
	Api PrometheusQueryAPI
}

func (p *PrometheusClient) QueryVectorValues(promql string) ([]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	value, err := p.Api.Query(ctx, promql, time.Time{})
	if err != nil {
		return nil, err
	}

	vector, ok := value.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("could not assert result from prometheus as Vector: %T", value)
	}

	values := make([]float64, 0, len(vector))
	for _, v := range vector {
		values = append(values, float64(v.Value))
	}
	return values, nil
}

func (p *PrometheusClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	return p.Api.Query(ctx, query, ts)
}
