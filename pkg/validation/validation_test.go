package validation_test

import (
	"testing"
	. "github.com/onsi/gomega"

	"net/http"
	"fmt"
	"time"
	"github.com/golang/protobuf/jsonpb"

	"code.cloudfoundry.org/cf-indicators/pkg/validation"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"

	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"bytes"
	"io/ioutil"
)

func TestVerifyMetric(t *testing.T) {

	t.Run("returns a result", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{func(req *http.Request) (*http.Response, error) {
			body := ioutil.NopCloser(bytes.NewBuffer(logCachePromQLResponse(3, 4)))
			return &http.Response{
				StatusCode:       200,
				Header:           http.Header(map[string][]string{"Content-Type": {"application/json"}}),
				Body:             body,
			}, nil
		}}

		m := indicator.Metric{
			Title:       "Demo Component",
			Origin:      "demo",
			SourceID:    "demo_component",
			Name:        "latency",
			Type:        "gauge",
			Description: "A test metric",
		}

		_, err := validation.VerifyMetric(m, `latency{source_id="demo_component",deployment="cf"}[1m]`, "http://server/v1/promql", client)

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("returns an error when the request fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{func(req *http.Request) (*http.Response, error) {
			return &http.Response{}, fmt.Errorf("oh no! can't get a port!")
		}}

		m := indicator.Metric{
			Title:       "Demo Component",
			Origin:      "demo",
			SourceID:    "demo_component",
			Name:        "latency",
			Type:        "gauge",
			Description: "A test metric",
		}

		_, err := validation.VerifyMetric(m, `latency{source_id="demo_component",deployment="cf"}[1m]`, "http://server/v1/promql", client)

		g.Expect(err).To(HaveOccurred())
	})

	t.Run("returns an error when the request is not 2xx", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:       401,
			}, nil
		}}

		m := indicator.Metric{
			Title:       "Demo Component",
			Origin:      "demo",
			SourceID:    "demo_component",
			Name:        "latency",
			Type:        "gauge",
			Description: "A test metric",
		}

		_, err := validation.VerifyMetric(m, `latency{source_id="demo_component",deployment="cf"}[1m]`, "http://server/v1/promql", client)

		g.Expect(err).To(HaveOccurred())
	})

	t.Run("returns an error when the response can not be unmarshalled", func(t *testing.T) {
		g := NewGomegaWithT(t)

		client := mockQueryClient{func(req *http.Request) (*http.Response, error) {
			body := ioutil.NopCloser(bytes.NewBuffer([]byte("will not unmarshal")))
			return &http.Response{
				StatusCode:       200,
				Header:           http.Header(map[string][]string{"Content-Type": {"application/json"}}),
				Body:             body,
			}, nil
		}}

		m := indicator.Metric{
			Title:       "Demo Component",
			Origin:      "demo",
			SourceID:    "demo_component",
			Name:        "latency",
			Type:        "gauge",
			Description: "A test metric",
		}

		_, err := validation.VerifyMetric(m, `latency{source_id="demo_component",deployment="cf"}[1m]`, "http://server/v1/promql", client)

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

func logCachePromQLResponse(numSeries, numPoints int) []byte {
	var series *logcache_v1.PromQL_Series
	var seriesList []*logcache_v1.PromQL_Series
	for i := 0; i < numSeries; i++ {
		series = &logcache_v1.PromQL_Series{
			Metric: map[string]string{
				"vm": fmt.Sprintf("vm-%d", i),
			},
		}

		series.Points = make([]*logcache_v1.PromQL_Point, numPoints)
		for j := 0; j < numPoints; j++ {
			series.Points[j] = &logcache_v1.PromQL_Point{
				Value: float64(j * i),
				Time:  time.Now().UnixNano(),
			}
		}

		seriesList = append(seriesList, series)
	}

	promQLResult := &logcache_v1.PromQL_QueryResult{
		Result: &logcache_v1.PromQL_QueryResult_Matrix{
			Matrix: &logcache_v1.PromQL_Matrix{
				Series: seriesList,
			},
		},
	}

	marshaller := jsonpb.Marshaler{}

	result, err := marshaller.MarshalToString(promQLResult)
	if err != nil {
		panic(err)
	}

	return []byte(result)
}

type mockQueryClient struct {
	TestDoer func(req *http.Request) (*http.Response, error)
}

func (m mockQueryClient) Do(req *http.Request) (*http.Response, error) {
	return m.TestDoer(req)
}
