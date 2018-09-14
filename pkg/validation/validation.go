package validation

import (
	"context"
	"fmt"
	"strings"
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

func FormatQuery(m indicator.Metric, deployment, lookback string) string {
	name := m.Name
	name = strings.Replace(name, `.`, "_", -1)
	name = strings.Replace(name, `-`, "_", -1)
	name = strings.Replace(name, `\`, "_", -1)
	name = strings.Replace(name, `/`, "_", -1)
	return fmt.Sprintf(`%s{source_id="%s",deployment="%s"}[%s]`, name, m.SourceID, deployment, lookback)
}

type promQLClient interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

func VerifyMetric(m indicator.Metric, query string, client promQLClient) (Result, error) {
	value, err := client.Query(context.Background(), query, time.Time{})

	if err != nil {
		return Result{}, fmt.Errorf("query failed [metric: %s] [query: %s] [status: %s]", m.Name, query, err)
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
