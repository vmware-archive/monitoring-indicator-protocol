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

	if value.Type() != model.ValMatrix {
		return Result{}, fmt.Errorf("query returned non-matrix result\n")
	}

	v := value.(model.Matrix)

	return buildResult(v), nil
}

func buildResult(queryResult model.Matrix) Result {

	var maxPoints int
	rs := make([]ResultSeries, 0)

	for _, s := range queryResult {
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

	return Result{
		MaxNumberOfPoints: maxPoints,
		Series:            rs,
	}
}
