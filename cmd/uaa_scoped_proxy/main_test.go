package main_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"os"
	"net/http"
	"code.cloudfoundry.org/indicators/pkg/go_test"
	"io/ioutil"
	"github.com/onsi/gomega/ghttp"
	"net/url"
	"time"
	"fmt"
	"net"
)

func TestUaaScopedProxy(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("passes authorized requests to the backend server", func(t *testing.T) {
		g := NewGomegaWithT(t)

		uaaServer := ghttp.NewServer()
		defer uaaServer.Close()
		uaaServer.AppendHandlers(
			ghttp.RespondWith(200, `{"scope":["abc-123","notifications.write"]}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		backendServer := ghttp.NewServer()
		defer backendServer.Close()
		backendServer.AppendHandlers(
			ghttp.RespondWith(200, `{"success":true}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		cmd := exec.Command(binPath,
			"-backend-url", "http://"+backendServer.Addr(),
			"-listen-addr", ":8080",
			"-uaa-url", "http://"+uaaServer.Addr(),
			"-uaa-ca-path", "./test_fixtures/ca.crt",
			"-log-cache-client", "my-uaa-client",
			"-log-cache-client-secret", "my-uaa-secret",
			"-k")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()
		waitForHTTPServer("localhost:8080", 3 * time.Second)

		url, err := url.Parse("http://localhost:8080")
		g.Expect(err).ToNot(HaveOccurred())
		req := &http.Request{
			Method:           http.MethodGet,
			URL:              url,
			Header:           map[string][]string{"Authorization":{"bearer abc-123"}},
		}
		resp, err := http.DefaultClient.Do(req)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		g.Eventually(uaaServer.ReceivedRequests).Should(HaveLen(1))
		g.Eventually(backendServer.ReceivedRequests).Should(HaveLen(1))

		bytes, err := ioutil.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bytes).To(ContainSubstring("{\"success\":true}"))
	})

	t.Run("rejects unauthorized requests with 401", func(t *testing.T) {
		g := NewGomegaWithT(t)

		uaaServer := ghttp.NewServer()
		defer uaaServer.Close()
		uaaServer.AppendHandlers(
			ghttp.RespondWith(200, `{"scope":["abc-123","nope.write"]}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		backendServer := ghttp.NewServer()
		defer backendServer.Close()
		backendServer.AppendHandlers(
			ghttp.RespondWith(200, `{"success":true}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		cmd := exec.Command(binPath,
			"-backend-url", "http://"+backendServer.Addr(),
			"-listen-addr", ":8081",
			"-uaa-url", "http://"+uaaServer.Addr(),
			"-uaa-ca-path", "./test_fixtures/ca.crt",
			"-log-cache-client", "my-uaa-client",
			"-log-cache-client-secret", "my-uaa-secret",
			"-k")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()
		waitForHTTPServer("localhost:8081", 3 * time.Second)

		url, err := url.Parse("http://localhost:8081")
		g.Expect(err).ToNot(HaveOccurred())
		req := &http.Request{
			Method:           http.MethodGet,
			URL:              url,
			Header:           map[string][]string{"Authorization":{"bearer abc-123"}},
		}
		resp, err := http.DefaultClient.Do(req)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

		g.Eventually(uaaServer.ReceivedRequests).Should(HaveLen(1))
		g.Consistently(backendServer.ReceivedRequests).Should(HaveLen(0))
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
