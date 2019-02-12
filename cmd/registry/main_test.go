package main_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal/indicator-protocol/pkg/configuration"
	"github.com/pivotal/indicator-protocol/pkg/go_test"
	"github.com/pivotal/indicator-protocol/pkg/mtls"
	"gopkg.in/yaml.v2"
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

	t.Run("it patches indicator documents when received", func(t *testing.T) {
		g := NewGomegaWithT(t)
		buffer := bytes.NewBuffer(nil)

		repoPath := go_test.CreateTempRepo("../../example_patch.yml", "../../example_indicators.yml")

		config := configuration.SourcesFile{
			Sources: []configuration.Source{{
				Type:       "git",
				Repository: repoPath,
				Glob:       "example_*.yml",
			}},
		}

		configBytes, err := yaml.Marshal(config)

		f, err := ioutil.TempFile("", "test_config.yml")
		_, err = f.Write(configBytes)
		g.Expect(err).ToNot(HaveOccurred())

		err = f.Close()
		g.Expect(err).ToNot(HaveOccurred())

		withConfigServer("10567", f.Name(), buffer, g, func(serverUrl string) {
			file, err := os.Open("test_fixtures/moar_indicators.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(resp.StatusCode, resp.Body).To(Equal(http.StatusOK))

			resp, err = client.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(buffer.String()).To(ContainSubstring("registered patch for name: my-other-component version: 1.2.3"))

			json, err := ioutil.ReadFile("test_fixtures/example_patched_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(responseBytes).To(MatchJSON(json))
		})
	})
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
