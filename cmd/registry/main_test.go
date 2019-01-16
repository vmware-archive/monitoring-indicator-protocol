package main_test

import (
	"testing"

	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/indicators/pkg/go_test"
	"code.cloudfoundry.org/indicators/pkg/mtls"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
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

	tlsConfig, err := mtls.NewClientConfig(clientCert, clientKey, rootCACert, "localhost")
	g.Expect(err).ToNot(HaveOccurred())

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	t.Run("it saves and exposes indicator documents", func(t *testing.T) {
		g := NewGomegaWithT(t)

		withServer("10567", g, func(serverUrl string) {
			file, err := os.Open("../../example.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(resp.StatusCode, resp.Body).To(Equal(http.StatusOK))

			resp, err = client.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			json, err := ioutil.ReadFile("../../pkg/registry/test_fixtures/example_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(len(json)).To(BeNumerically(">", 200))
			g.Expect(responseBytes).To(MatchJSON(json))
		})
	})

	t.Run("it loads patches from git sources", func(t *testing.T) {
		g := NewGomegaWithT(t)
		buffer := bytes.NewBuffer(nil)

		withConfigServer("10567", "test_fixtures/git_config.yml", buffer, g, func(serverUrl string) {
			results := buffer.String()
			g.Expect(results).To(ContainSubstring("registered patch for name: my-component version: 1.2.3"))
			g.Expect(results).To(ContainSubstring("registered patch for name: much-yaml-component version: 1.2.3"))
			g.Expect(results).ToNot(ContainSubstring("registered patch for\n"))
		})
	})

	t.Run("it loads indicator documents from git sources", func(t *testing.T) {
		g := NewGomegaWithT(t)
		buffer := bytes.NewBuffer(nil)

		withConfigServer("10567", "test_fixtures/git_config.yml", buffer, g, func(serverUrl string) {
			resp, err := client.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responseBytes).To(ContainSubstring("success_percentage"))
		})
	})

	t.Run("it loads documents and patches from git sources based on glob", func(t *testing.T) {
		g := NewGomegaWithT(t)
		buffer := bytes.NewBuffer(nil)

		withConfigServer("10567", "test_fixtures/git_glob_config.yml", buffer, g, func(serverUrl string) {
			resp, err := client.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			json, err := ioutil.ReadFile("../../pkg/registry/test_fixtures/example_patched_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(len(json)).To(BeNumerically(">", 200))
			g.Expect(responseBytes).To(MatchJSON(json))
		})
	})

	t.Run("it patches indicator documents when received", func(t *testing.T) {
		g := NewGomegaWithT(t)

		buffer := bytes.NewBuffer(nil)
		withConfigServer("10567", "test_fixtures/local_config.yml", buffer, g, func(serverUrl string) {
			file, err := os.Open("../../example.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(resp.StatusCode, resp.Body).To(Equal(http.StatusOK))

			resp, err = client.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(buffer.String()).To(ContainSubstring("registered patch for name: my-component version: 1.2.3"))

			json, err := ioutil.ReadFile("../../pkg/registry/test_fixtures/example_patched_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(len(json)).To(BeNumerically(">", 200))
			g.Expect(responseBytes).To(MatchJSON(json))
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
	buffer := bytes.NewBuffer(nil)
	session, err := gexec.Start(cmd, buffer, buffer)
	g.Expect(err).ToNot(HaveOccurred())
	defer session.Kill()
	serverHost := "localhost:" + port
	err = go_test.WaitForHTTPServer(serverHost, 3*time.Second)
	g.Expect(err).ToNot(HaveOccurred())
	testFun("https://" + serverHost)
}

func withConfigServer(port, configPath string, buffer *bytes.Buffer, g *GomegaWithT, testFun func(string)) {
	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	cmd := exec.Command(binPath,
		"--port", port,
		"--tls-pem-path", serverCert,
		"--tls-key-path", serverKey,
		"--tls-root-ca-pem", rootCACert,
		"--config", configPath,
	)

	session, err := gexec.Start(cmd, buffer, buffer)
	g.Expect(err).ToNot(HaveOccurred())
	defer session.Kill()
	serverHost := "localhost:" + port
	err = go_test.WaitForHTTPServer(serverHost, 3*time.Second)
	g.Expect(err).ToNot(HaveOccurred())
	testFun("https://" + serverHost)
}
