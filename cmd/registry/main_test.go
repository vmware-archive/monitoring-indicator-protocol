package main_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"os/exec"
	"os"
	"net/http"
	"io/ioutil"
	"time"
	"fmt"
	"net"

	"code.cloudfoundry.org/cf-indicators/pkg/go_test"
)

func TestIndicatorRegistry(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it saves and exposes indicators with labels", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath, "--port", "10567")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()
		waitForHTTPServer("localhost:10567", 3 * time.Second)

		file, err := os.Open("../../example.yml")
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := http.Post("http://localhost:10567/v1/register?deployment=redis-abc&service=redis", "text/plain", file)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		resp, err = http.Get("http://localhost:10567/v1/indicator-documents")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		bytes, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bytes).To(MatchJSON(`
			[
              {
                "labels": {
                  "deployment": "redis-abc",
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

	t.Run("it exposes a metrics endpoint", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath, "--port", "10568")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()
		waitForHTTPServer("localhost:10568", 3 * time.Second)

		resp, err := http.Get("http://localhost:10568/metrics")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("it records metrics for all endpoints", func(t *testing.T) {
		g := NewGomegaWithT(t)

		cmd := exec.Command(binPath, "--port", "10569")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()
		waitForHTTPServer("localhost:10569", 3 * time.Second)

		file, err := os.Open("../../example.yml")
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := http.Post("http://localhost:10569/v1/register?deployment=redis-abc&service=redis", "text/plain", file)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		resp, err = http.Get("http://localhost:10569/v1/indicator-documents")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		resp, err = http.Get("http://localhost:10569/v2/fake-endpoint")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

		resp, err = http.Get("http://localhost:10569/metrics")
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
