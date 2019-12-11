package main_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

const documentsResponseBody string = `[
  {
	"apiVersion": "v1",
	"uid": "my-other-component-c2dd9",
    "kind": "IndicatorDocument",
	"metadata": {
      "labels": {}
    },
    "spec" : {
	  "product": {
	    "name": "my-other-component",
	    "version": "1.2.3"
	  },
	  "indicators": [
	    {
		  "name": "very_good_indicator",
		  "promql": "avg_over_time(latency[5m])",
		  "thresholds": [
			  {"level": "critical", "operator": "gte", "value": 11},
			  {"level": "warning", "operator": "gt", "value": 9}
		  ],
		  "alert": {
		    "step": "10m"
		  }
	    }
	  ]
    }
  }
]`

const prometheusResponse string = `{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "deployment": "healthwatch-v1-5",
          "source_id": "healthwatch-forwarder"
        },
        "value": [
			1554999048.236,
            "10"
        ]
      }
    ]
  }
}
`

var (
	rootCACert = "../../test_fixtures/server.pem"
	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestStatusControllerBinary(t *testing.T) {
	t.Run("it updates indicator status for every document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		prometheusServer := setupFakePrometheusServer(g)
		defer prometheusServer.Close()

		registryServer := setupFakeRegistry(g)
		defer registryServer.Close()

		oauthServer := setupFakeOauthServer(g)
		defer oauthServer.Close()

		binPath, err := go_test.Build("./", "-race")
		g.Expect(err).ToNot(HaveOccurred())

		cmd := exec.Command(
			binPath,
			"--registry-uri", registryServer.URL(),
			"--prometheus-uri", prometheusServer.URL(),
			"--tls-pem-path", clientCert,
			"--tls-key-path", clientKey,
			"--tls-root-ca-pem", rootCACert,
			"--tls-server-cn", "localhost",
			"--oauth-server", oauthServer.URL(),
			"--oauth-client-id", "alana",
			"--oauth-client-secret", "abc123",
		)

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()

		g.Eventually(prometheusServer.ReceivedRequests, 2*time.Second).Should(HaveLen(1))
		g.Eventually(registryServer.ReceivedRequests, 2*time.Second).Should(HaveLen(2))
	})
}

func setupFakeRegistry(g *GomegaWithT) *ghttp.Server {
	registryServer := ghttp.NewServer()
	registryServer.AppendHandlers(
		func(w http.ResponseWriter, r *http.Request) {
			g.Expect(r.Method).To(Equal("GET"))

			g.Expect(r.URL.Path).To(Equal("/v1/indicator-documents"))
			body := []byte(documentsResponseBody)
			_, err := w.Write(body)
			g.Expect(err).NotTo(HaveOccurred())
		},
		func(w http.ResponseWriter, r *http.Request) {
			g.Expect(r.Method).To(Equal("POST"))
			g.Expect(r.URL.Path).To(Equal("/v1/indicator-documents/my-other-component-c2dd9/bulk_status"))
			var indicatorStatuses []registry.ApiV1UpdateIndicatorStatus
			err := json.NewDecoder(r.Body).Decode(&indicatorStatuses)
			g.Expect(err).NotTo(HaveOccurred())

			status := "warning"
			g.Expect(indicatorStatuses).To(BeEquivalentTo([]registry.ApiV1UpdateIndicatorStatus{{
				Name:   "very_good_indicator",
				Status: &status,
			}}))

			w.WriteHeader(http.StatusOK)
		},
	)
	return registryServer
}

func setupFakeOauthServer(g *GomegaWithT) *ghttp.Server {
	oauthServer := ghttp.NewServer()
	oauthServer.AppendHandlers(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"token_type": "bearer", "access_token": "my-token"}`))
		g.Expect(err).ToNot(HaveOccurred())
	})
	return oauthServer
}

func setupFakePrometheusServer(g *GomegaWithT) *ghttp.Server {
	prometheusServer := ghttp.NewServer()
	prometheusServer.AppendHandlers(func(w http.ResponseWriter, r *http.Request) {
		g.Expect(r.URL.Path).To(Equal(`/api/v1/query`))
		g.Expect(r.URL.Query()).To(BeEquivalentTo(url.Values{"query": []string{`avg_over_time(latency[5m])`}}))
		g.Expect(r.Header.Get("Authorization")).To(Equal("bearer my-token"))

		w.Header().Set("Content-Type", "application/json")
		body := []byte(prometheusResponse)
		_, err := w.Write(body)

		g.Expect(err).NotTo(HaveOccurred())
	})
	return prometheusServer
}
