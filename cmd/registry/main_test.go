package main_test

import (
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"testing"

	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/indicators/pkg/go_test"
	"code.cloudfoundry.org/indicators/pkg/mtls"
)

var (
	serverCert = "../../test_fixtures/leaf.pem"
	serverKey  = "../../test_fixtures/leaf.key"
	rootCACert = "../../test_fixtures/root.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestIndicatorRegistry(t *testing.T) {
	g := NewGomegaWithT(t)

	client, err := mtls.NewClient(clientCert, clientKey, rootCACert)
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it saves and exposes indicators with labels", func(t *testing.T) {
		g := NewGomegaWithT(t)

		withServer("10567", g, func(serverUrl string) {
			file, err := os.Open("../../example.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(serverUrl+"/v1/register?deployment=redis-abc&service=redis", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = client.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			bytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(bytes).To(MatchJSON(`
				  [
			        {
			          "labels": {
			            "deployment": "redis-abc",
			            "product": "my-component",
			            "service": "redis"
			          },
			          "indicators": [
			            {
			              "name": "doc_performance_indicator",
			              "title": "Doc Performance Indicator",
			              "metrics": [
			                {
			                  "name": "latency",
			                  "source_id": "demo",
			                  "origin": "demo",
			                  "title": "Demo Latency",
			                  "type": "gauge",
			                  "frequency": "1m",
			                  "description": "A test metric for testing"
			                }
			              ],
			              "measurement": "Average latency over last 5 minutes per instance",
			              "promql": "avg_over_time(demo_latency{source_id=\"doc\"}[5m])",
			              "thresholds": [
			                {
			                "level": "warning",
			                "dynamic": true,
			                "operator": "gte",
			                "value": 50
			              },
			              {
			                "level": "critical",
			                "dynamic": true,
			                "operator": "gte",
			                "value": 100
			              }
			              ],
			              "description": "This is a valid markdown description.\n\n**Use**: This indicates nothing. It is placeholder text.\n\n**Type**: Gauge\n**Frequency**: 60 s\n",
			              "response": "Panic! Run around in circles flailing your arms.\n"
			            }
			          ]
			        }
		          ]  
				`))
		})
	})

	t.Run("it exposes a metrics endpoint", func(t *testing.T) {
		g := NewGomegaWithT(t)
		withServer("10568", g, func(serverUrl string) {
			resp, err := client.Get(serverUrl + "/metrics")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	t.Run("it records metrics for all endpoints", func(t *testing.T) {
		g := NewGomegaWithT(t)

		withServer("10569", g, func(serverUrl string) {
			file, err := os.Open("../../example.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(serverUrl+"/v1/register?deployment=redis-abc&service=redis", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = client.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = client.Get(serverUrl + "/v2/fake-endpoint")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

			resp, err = client.Get(serverUrl + "/metrics")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
			respBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			respString := string(respBytes)
			g.Expect(respString).To(ContainSubstring(`registry_http_requests{route="/v1/indicator-documents",status="200"} 1`))
			g.Expect(respString).To(ContainSubstring(`registry_http_requests{route="/v1/register",status="200"} 1`))
			g.Expect(respString).To(ContainSubstring(`registry_http_requests{route="invalid path",status="404"} 1`))
		})
	})

	t.Run("it fails tls handshake with bad certs", func(t *testing.T) {
		g := NewGomegaWithT(t)

		withServer("10570", g, func(serverUrl string) {
			g.Expect(err).ToNot(HaveOccurred())

			badClient := http.Client{
				Transport: nil,
			}

			_, err = badClient.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).To(HaveOccurred())
		})
	})
}

func withServer(port string, g *GomegaWithT, testFun func(string)) {
	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	cmd := exec.Command(binPath,
		"--port", port,
		"--tls-pem-path", serverCert,
		"--tls-key-path", serverKey,
		"--tls-root-ca-pem", rootCACert,
	)
	session, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	g.Expect(err).ToNot(HaveOccurred())
	defer session.Kill()
	serverHost := "localhost:" + port
	waitForHTTPServer(serverHost, 3*time.Second)
	testFun("https://" + serverHost)
}

func waitForHTTPServer(host string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)

	for {
		select {
		case <-timer.C:
			return fmt.Errorf("http server [%s] did not start", host)
		default:
			_, err := net.DialTimeout("tcp", host, 50*time.Millisecond)
			if err == nil {
				return nil
			}
		}
	}
}
