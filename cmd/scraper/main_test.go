package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
)

type proxyConfig struct {
	// Filenames of the server and client keys and certs
	ServerKey  string
	ServerCert string
	ClientKey  string
	ClientCert string
	CaCert     string

	RegistryPort int
	ProxyPort    int
}

func TestScraper(t *testing.T) {
	g := NewGomegaWithT(t)

	localServerCert := "../../test_fixtures/server.pem"
	localServerKey := "../../test_fixtures/server.key"
	localRootCACert := "../../test_fixtures/ca.pem"
	localClientKey := "../../test_fixtures/client.key"
	localClientCert := "../../test_fixtures/client.pem"

	localTlsConfig, err := mtls.NewClientConfig(localClientCert, localClientKey, localRootCACert, "localhost")
	g.Expect(err).ToNot(HaveOccurred())

	localClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: localTlsConfig,
		},
	}
	remoteTlsConfig, err := mtls.NewClientConfig(localClientCert, localClientKey, localRootCACert, "localhost")
	g.Expect(err).ToNot(HaveOccurred())

	remoteClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: remoteTlsConfig,
		},
	}

	t.Run("It forwards documents with additional metadata", func(t *testing.T) {
		g := NewGomegaWithT(t)
		// TODO use different certs (from different CAs) for the configurations
		localConfig := proxyConfig{
			ServerKey:    localServerKey,
			ServerCert:   localServerCert,
			ClientKey:    localClientKey,
			ClientCert:   localClientCert,
			CaCert:       localRootCACert,
			RegistryPort: 19281,
			ProxyPort:    19282,
		}

		remoteConfig := proxyConfig{
			ServerKey:    localServerKey,
			ServerCert:   localServerCert,
			ClientKey:    localClientKey,
			ClientCert:   localClientCert,
			CaCert:       localRootCACert,
			RegistryPort: 19283,
			ProxyPort:    19284,
		}

		localProxySession, localRegistrySession := startProxy(g, localConfig)
		remoteProxySession, remoteRegistrySession := startProxy(g, remoteConfig)

		defer localProxySession.Kill()
		defer localRegistrySession.Kill()
		defer remoteProxySession.Kill()
		defer remoteRegistrySession.Kill()

		scraperSession, err := startScraper(g, localConfig, remoteConfig)
		g.Expect(err).ToNot(HaveOccurred())
		defer scraperSession.Kill()

		// Add a document to the remote registry
		file, err := os.Open("test_fixtures/indicators.yml")
		g.Expect(err).ToNot(HaveOccurred())
		createDocumentUrl := fmt.Sprintf("https://localhost:%d/v1/register", remoteConfig.ProxyPort)
		resp, err := remoteClient.Post(createDocumentUrl, "application/yml", file)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		// Observe that the document was scraped and put in the local registry, with some metadata added
		getDocumentsUrl := fmt.Sprintf("https://localhost:%d/v1/indicator-documents", localConfig.ProxyPort)
		g.Expect(err).NotTo(HaveOccurred())
		g.Eventually(func() ([]v1.IndicatorDocument, error) {
			response1, err := localClient.Get(getDocumentsUrl)
			if err != nil {
				return nil, err
			}
			responseBytes, err := ioutil.ReadAll(response1.Body)
			var docs []v1.IndicatorDocument
			err = yaml.Unmarshal(responseBytes, &docs)
			return docs, err
		}).Should(HaveLen(1))

		response1, err := localClient.Get(getDocumentsUrl)
		g.Expect(err).ToNot(HaveOccurred())
		responseBytes, err := ioutil.ReadAll(response1.Body)
		expectedJSON, err := ioutil.ReadFile("test_fixtures/response.json")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(responseBytes).Should(MatchJSON(expectedJSON))

	})
}

func startProxy(g *GomegaWithT, config proxyConfig) (*gexec.Session, *gexec.Session) {
	binPath, err := go_test.Build("../registry_proxy", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	cmd := exec.Command(
		binPath,
		"--port", strconv.Itoa(config.ProxyPort),
		"--tls-pem-path", config.ServerCert,
		"--tls-client-pem-path", config.ClientCert,
		"--tls-key-path", config.ServerKey,
		"--tls-client-key-path", config.ClientKey,
		"--tls-root-ca-pem", config.CaCert,
		"--local-registry-addr", fmt.Sprintf("localhost:%d", config.RegistryPort),
		"--tls-server-cn", "localhost",
	)
	proxySession, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	g.Expect(err).ToNot(HaveOccurred())

	err = go_test.WaitForTCPServer(fmt.Sprintf("localhost:%d", config.ProxyPort), 3*time.Second)
	g.Expect(err).ToNot(HaveOccurred())

	binPath, err = go_test.Build("../registry", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	cmd = exec.Command(
		binPath,
		"--port", strconv.Itoa(config.RegistryPort),
	)
	registrySession, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	g.Expect(err).ToNot(HaveOccurred())

	err = go_test.WaitForTCPServer(fmt.Sprintf("localhost:%d", config.RegistryPort), 3*time.Second)
	g.Expect(err).ToNot(HaveOccurred())

	return proxySession, registrySession
}

func startScraper(g *GomegaWithT, localConfig proxyConfig, remoteConfig proxyConfig) (*gexec.Session, error) {
	binPath, err := go_test.Build("./", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	cmd := exec.Command(
		binPath,
		"--interval", "1s",
		"--local-key-path", localConfig.ClientKey,
		"--remote-key-path", remoteConfig.ClientKey,
		"--local-pem-path", localConfig.ClientCert,
		"--remote-pem-path", remoteConfig.ClientCert,
		"--local-root-ca-pem", localConfig.CaCert,
		"--remote-root-ca-pem", remoteConfig.CaCert,
		"--local-registry-addr", fmt.Sprintf("localhost:%d", localConfig.RegistryPort),
		"--remote-registry-addr", fmt.Sprintf("localhost:%d", remoteConfig.RegistryPort),
		"--local-server-cn", "localhost",
		"--remote-server-cn", "localhost",
	)

	scraperSession, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	return scraperSession, err
}
