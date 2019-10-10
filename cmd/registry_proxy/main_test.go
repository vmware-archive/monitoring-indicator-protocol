package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/tls_config"
)

var (
	serverCert = "../../test_fixtures/server.pem"
	serverKey  = "../../test_fixtures/server.key"
	rootCACert = "../../test_fixtures/ca.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestIndicatorRegistryProxy(t *testing.T) {
	g := NewGomegaWithT(t)
	tlsConfig, err := tls_config.NewClientConfig(clientCert, clientKey, rootCACert, "localhost")
	g.Expect(err).ToNot(HaveOccurred())

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	t.Run("it creates handlers for all registries", func(t *testing.T) {
		g := NewGomegaWithT(t)
		registry1Port := "23456"
		registry2Port := "45678"
		proxy1Port := "10569"
		proxy2Port := "10568"
		session1, regSession1 := startSession(g, proxy1Port, registry1Port, proxy2Port)
		session2, regSession2 := startSession(g, proxy2Port, registry2Port, proxy1Port)

		defer session1.Kill()
		defer regSession1.Kill()
		defer session2.Kill()
		defer regSession2.Kill()

		createDocumentUrl := fmt.Sprintf("https://localhost:%s/v1/register", proxy1Port)
		getDocumentsUrl := "https://localhost:%s/v1/indicator-documents"

		file, err := os.Open("test_fixtures/indicators.yml")
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Post(createDocumentUrl, "application/yml", file)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		response1, err := client.Get(fmt.Sprintf(getDocumentsUrl, proxy1Port))
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response1.StatusCode).To(Equal(http.StatusOK))
		response1Bytes, err := ioutil.ReadAll(response1.Body)
		expectedJSON, err := ioutil.ReadFile("test_fixtures/response.json")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(response1Bytes).Should(MatchJSON(expectedJSON))

		response2, err := client.Get(fmt.Sprintf(getDocumentsUrl, proxy2Port))
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response2.StatusCode).To(Equal(http.StatusOK))
		response2Bytes, err := ioutil.ReadAll(response2.Body)

		g.Expect(response2Bytes).Should(MatchJSON(expectedJSON))
	})
}

func startSession(g *GomegaWithT, port string, localRegistryPort string, otherProxyPort string) (*gexec.Session, *gexec.Session) {
	binPath, err := go_test.Build("./", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	cmd := exec.Command(
		binPath,
		"--port", port,
		"--tls-pem-path", serverCert,
		"--tls-client-pem-path", clientCert,
		"--tls-key-path", serverKey,
		"--tls-client-key-path", clientKey,
		"--tls-root-ca-pem", rootCACert,
		"--local-registry-addr", fmt.Sprintf("localhost:%s", localRegistryPort),
		"--registry-addr", fmt.Sprintf("localhost:%s", otherProxyPort),
		"--tls-server-cn", "localhost",
	)
	proxySession, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	g.Expect(err).ToNot(HaveOccurred())

	err = go_test.WaitForTCPServer(fmt.Sprintf("localhost:%s", port), 3*time.Second)
	g.Expect(err).ToNot(HaveOccurred())

	binPath, err = go_test.Build("../registry", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	cmd = exec.Command(
		binPath,
		"--port", localRegistryPort,
	)
	registrySession, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	g.Expect(err).ToNot(HaveOccurred())

	err = go_test.WaitForTCPServer(fmt.Sprintf("localhost:%s", localRegistryPort), 3*time.Second)
	g.Expect(err).ToNot(HaveOccurred())

	return proxySession, registrySession
}
