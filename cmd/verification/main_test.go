package main_test

import (
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"testing"

	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/indicators/pkg/go_test"

	"github.com/prometheus/common/model"
)

func TestValidateIndicators(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("returns 0 when all indicators return data", func(t *testing.T) {
		g := NewGomegaWithT(t)

		logCacheServer := ghttp.NewServer()
		defer logCacheServer.Close()

		logCacheServer.AppendHandlers(
			func(w http.ResponseWriter, req *http.Request) {
				req.ParseForm()

				q := req.Form.Get("query")
				if q != `avg_over_time(demo_latency{source_id="demo_component",deployment="my-demo-deployment"}[5m])` {
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

				if q != `saturation{source_id="demo_component",deployment="my-demo-deployment"}` {
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
			"--metadata", "deployment=my-demo-deployment",
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

	t.Run("returns 1 when not all indicators return data", func(t *testing.T) {
		g := NewGomegaWithT(t)

		logCacheServer := ghttp.NewServer()
		defer logCacheServer.Close()

		logCacheServer.AppendHandlers(
			func(w http.ResponseWriter, req *http.Request) {
				req.ParseForm()
				q := req.Form.Get("query")
				g.Expect(q).To(Equal(`avg_over_time(demo_latency{source_id="demo_component",deployment="cf"}[5m])`))

				body := logCachePromQLResponse(3, 4)
				w.Write(body)
				w.Header().Set("Content-Type", "application/json")
			},
			func(w http.ResponseWriter, req *http.Request) {
				req.ParseForm()
				q := req.Form.Get("query")
				g.Expect(q).To(Equal(`saturation{source_id="demo_component",deployment="cf"}`))

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
			"--metadata", "deployment=cf",
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
	var series *model.SampleStream
	var seriesList model.Matrix
	for i := 0; i < numSeries; i++ {
		series = &model.SampleStream{
			Metric: model.Metric{
				"vm": model.LabelValue(fmt.Sprintf("vm-%d", i)),
			},
			Values: make([]model.SamplePair, numPoints),
		}

		for j := 0; j < numPoints; j++ {
			series.Values[j] = model.SamplePair{
				Value:     model.SampleValue(float64(j * i)),
				Timestamp: model.Time(time.Now().Unix()),
			}
		}

		seriesList = append(seriesList, series)
	}

	result, err := json.Marshal(response{
		"success",
		data{
			ResultType: "matrix",
			Result:     seriesList,
		},
	})

	if err != nil {
		panic(err)
	}

	return result
}

type response struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string      `json:"resultType"`
	Result     interface{} `json:"result"`
}
