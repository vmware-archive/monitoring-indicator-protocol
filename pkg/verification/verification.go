package verification

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"

	"code.cloudfoundry.org/indicators/pkg/indicator"
)

type Result struct {
	MaxNumberOfPoints int
	Series            []ResultSeries
}

type ResultSeries struct {
	Labels string
	Points []string
}

type promQLClient interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

func VerifyIndicator(i indicator.Indicator, client promQLClient) (Result, error) {
	value, err := client.Query(context.Background(), i.PromQL, time.Time{})

	if err != nil {
		return Result{}, fmt.Errorf("query failed [indicator: %s] [promql: %s] [status: %s]", i.Name, i.PromQL, err)
	}

	return buildResult(value)
}

func buildResult(queryResult model.Value) (Result, error) {

	var maxPoints int
	rs := make([]ResultSeries, 0)

	switch queryResult.Type() {

	case model.ValMatrix:
		for _, s := range queryResult.(model.Matrix) {
			ps := make([]string, 0)
			for _, p := range s.Values {
				ps = append(ps, p.Value.String())
			}

			rs = append(rs, ResultSeries{
				Labels: s.Metric.String(),
				Points: ps,
			})

			if len(ps) > maxPoints {
				maxPoints = len(ps)
			}
		}

	case model.ValVector:
		qr := queryResult.(model.Vector)
		ps := make([]string, 0)
		for _, p := range qr {
			ps = append(ps, p.Value.String())
		}

		labels := ""
		if len(qr) > 0 {
			labels = qr[0].String()
		}
		rs = append(rs, ResultSeries{
			Labels: labels,
			Points: ps,
		})

		maxPoints = len(ps)

	default:
		return Result{}, fmt.Errorf("result type not supported: %s", queryResult.Type())
	}

	return Result{
		MaxNumberOfPoints: maxPoints,
		Series:            rs,
	}, nil
}
