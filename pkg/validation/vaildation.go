package validation

import (
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"code.cloudfoundry.org/go-log-cache"
	"net/http"
	"strings"
	"fmt"
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"github.com/golang/protobuf/jsonpb"
)

type Result struct {
	MaxNumberOfPoints int
	Series            []ResultSeries
}

type ResultSeries struct {
	Labels map[string]string
	Points []float64
}

func FormatQuery(m indicator.Metric, deployment string) string {
	name := strings.Replace(m.Name, ".", "_", -1)
	return fmt.Sprintf(`%s{source_id="%s",deployment="%s"}[1m]`, name, m.SourceID, deployment)
}

func VerifyMetric(m indicator.Metric, query string, logCacheURL string, logCache *logcache.Oauth2HTTPClient) (Result, error) {
	req, err := http.NewRequest(http.MethodGet, logCacheURL+"/v1/promql", nil)
	if err != nil {
		return Result{}, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := logCache.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("failed to query log-cache [metric: %s] [query: %s] [error: %s]", m.Name, query, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Result{}, fmt.Errorf("log-cache query returned bad status [metric: %s] [query: %s] [status: %d]", m.Name, query, resp.StatusCode)
	}

	defer resp.Body.Close()

	var r logcache_v1.PromQL_QueryResult
	err = jsonpb.Unmarshal(resp.Body, &r)
	if err != nil {
		return Result{}, fmt.Errorf("failed to unmarshal log-cache query response [metric: %s] [query: %s] [error: %s]", m.Name, query, err)
	}

	return buildResult(r), nil
}

func buildResult(queryResult logcache_v1.PromQL_QueryResult) Result {
	var maxPoints int
	rs := make([]ResultSeries, 0)

	for _, s := range queryResult.GetMatrix().Series {
		ps := make([]float64, 0)
		for _, p := range s.Points {
			ps = append(ps, p.Value)
		}

		rs = append(rs, ResultSeries{
			Labels: s.GetMetric(),
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
