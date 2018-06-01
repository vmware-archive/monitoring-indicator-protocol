package main_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	"os/exec"
	"os"
	"net/http"
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"github.com/golang/protobuf/jsonpb"
	"time"
	"fmt"
)

func TestValidateIndicators(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := gexec.Build("./main.go")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("returns 0 when all metrics are found over 1m", func(t *testing.T) {
		g := NewGomegaWithT(t)

		logCacheServer := ghttp.NewServer()
		defer logCacheServer.Close()

		logCacheServer.AppendHandlers(
			func(w http.ResponseWriter, req *http.Request) {
				req.ParseForm()

				q := req.Form.Get("query")
				if q != `latency{source_id="demo_component",deployment="cf"}[1m]` {
					w.WriteHeader(422)
					return
				}

				body := logCachePromQLResponse(3, 4)
				w.Write(body)
				w.Header().Set("Content-Type", "application/json")
			},
			func(w http.ResponseWriter, req *http.Request) {
				req.ParseForm()
				q := req.Form.Get("query")
				if q != `saturation{source_id="demo_component",deployment="cf"}[1m]` {
					w.WriteHeader(422)
					return
				}

				body := logCachePromQLResponse(1, 1)
				w.Write(body)
				w.Header().Set("Content-Type", "application/json")
			},
		)

		uaaServer := ghttp.NewServer()
		defer uaaServer.Close()
		uaaServer.AppendHandlers(
			ghttp.RespondWith(200, `{"access_token":"abc-123"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		cmd := exec.Command(
			binPath,
			"--indicators", "./test_fixtures/indicators.yml",
			"--deployment", "cf",
			"--log-cache-url", "http://"+logCacheServer.Addr(),
			"--uaa-url", "http://"+uaaServer.Addr(),
			"--log-cache-client", "my-uaa-client",
			"--log-cache-client-secret", "client-secret",
			"-k",
		)

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)
		g.Expect(err).ToNot(HaveOccurred())

		g.Eventually(uaaServer.ReceivedRequests).Should(HaveLen(1))
		g.Eventually(logCacheServer.ReceivedRequests).Should(HaveLen(2))
		g.Eventually(session).Should(gexec.Exit(0))
	})

	t.Run("returns 1 when not all metrics are found over 1m", func(t *testing.T) {
		g := NewGomegaWithT(t)

		logCacheServer := ghttp.NewServer()
		defer logCacheServer.Close()

		logCacheServer.AppendHandlers(
			func(w http.ResponseWriter, req *http.Request) {
				req.ParseForm()

				q := req.Form.Get("query")
				if q != `latency{source_id="demo_component",deployment="cf"}[1m]` {
					w.WriteHeader(422)
					return
				}

				body := logCachePromQLResponse(3, 4)
				w.Write(body)
				w.Header().Set("Content-Type", "application/json")
			},
			func(w http.ResponseWriter, req *http.Request) {
				req.ParseForm()
				q := req.Form.Get("query")
				if q != `saturation{source_id="demo_component",deployment="cf"}[1m]` {
					w.WriteHeader(422)
					return
				}

				body := logCachePromQLResponse(0, 0)
				w.Write(body)
				w.Header().Set("Content-Type", "application/json")
			},
		)

		uaaServer := ghttp.NewServer()
		defer uaaServer.Close()
		uaaServer.AppendHandlers(
			ghttp.RespondWith(200, `{"access_token":"abc-123"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		cmd := exec.Command(
			binPath,
			"--indicators", "./test_fixtures/indicators.yml",
			"--deployment", "cf",
			"--log-cache-url", "http://"+logCacheServer.Addr(),
			"--uaa-url", "http://"+uaaServer.Addr(),
			"--log-cache-client", "my-uaa-client",
			"--log-cache-client-secret", "client-secret",
			"-k",
		)

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)
		g.Expect(err).ToNot(HaveOccurred())

		g.Eventually(uaaServer.ReceivedRequests).Should(HaveLen(1))
		g.Eventually(logCacheServer.ReceivedRequests).Should(HaveLen(2))
		g.Eventually(session).Should(gexec.Exit(1))
	})
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
