package main_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/benjamintf1/unmarshalledmatchers"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/configuration"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
)

func TestIndicatorRegistry(t *testing.T) {
	t.Run("it patches indicator documents when received", func(t *testing.T) {
		g := NewGomegaWithT(t)

		repoPath := go_test.CreateTempRepo("../../example_patch_add_indicator.yml", "../../example_indicators.yml")

		config := configuration.SourcesFile{
			Sources: []configuration.Source{{
				Type:       "git",
				Repository: repoPath,
				Glob:       "example_*.yml",
			}},
		}

		configBytes, err := yaml.Marshal(config)

		f, err := ioutil.TempFile("", "test_config.yml")
		g.Expect(err).ToNot(HaveOccurred())
		_, err = f.Write(configBytes)
		g.Expect(err).ToNot(HaveOccurred())

		err = f.Close()
		g.Expect(err).ToNot(HaveOccurred())

		withConfigServer("10567", f.Name(), g, func(serverUrl string) {
			file, err := os.Open("test_fixtures/indicators.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := http.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(resp.StatusCode, resp.Body).To(Equal(http.StatusOK))

			resp, err = http.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			expectedJSON, err := ioutil.ReadFile("test_fixtures/patched_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(responseBytes).To(MatchJSON(expectedJSON))
		})
	})

	t.Run("it saves indicator status", func(t *testing.T) {
		g := NewGomegaWithT(t)

		withConfigServer("10567", "", g, func(serverUrl string) {
			file, err := os.Open("test_fixtures/indicators.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := http.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			file, err = os.Open("test_fixtures/bulk_status_request.json")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = http.Post(serverUrl+"/v1/indicator-documents/my-other-component-62a5511746dfd09059ced03b2ed73ff0ae942421/bulk_status", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = http.Get(serverUrl + "/v1/indicator-documents")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			expectedJSON, err := ioutil.ReadFile("test_fixtures/status_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(responseBytes).Should(ContainOrderedJSON(expectedJSON))
		})
	})

	t.Run("it retrieves documents by product name", func(t *testing.T) {
		g := NewGomegaWithT(t)

		withConfigServer("1093", "", g, func(serverUrl string) {
			file, err := os.Open("test_fixtures/indicators.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := http.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			file, err = os.Open("test_fixtures/indicators2.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = http.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = http.Get(serverUrl + "/v1/indicator-documents?product-name=my-other-other-component")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			expectedJSON, err := ioutil.ReadFile("test_fixtures/filtered_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(responseBytes).Should(ContainOrderedJSON(expectedJSON))

		})
	})

	t.Run("it retrieves documents by any metadata key", func(t *testing.T) {
		g := NewGomegaWithT(t)

		withConfigServer("1093", "", g, func(serverUrl string) {
			file, err := os.Open("test_fixtures/indicators.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := http.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			file, err = os.Open("test_fixtures/indicators2.yml")
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = http.Post(serverUrl+"/v1/register", "text/plain", file)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = http.Get(serverUrl + "/v1/indicator-documents?deployment=my-service-deployment-2")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBytes, err := ioutil.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())

			expectedJSON, err := ioutil.ReadFile("test_fixtures/filtered_response.json")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(responseBytes).Should(ContainOrderedJSON(expectedJSON))

		})
	})
}

func withConfigServer(port, configPath string, g *GomegaWithT, testFun func(string)) {
	binPath, err := go_test.Build("./", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	cmd := exec.Command(
		binPath,
		"--port", port,
		"--config", configPath,
	)

	var outW, errW io.Writer
	if testing.Verbose() {
		outW = os.Stdout
		errW = os.Stderr
	}
	session, err := gexec.Start(cmd, outW, errW)
	g.Expect(err).ToNot(HaveOccurred())
	defer session.Kill()
	serverHost := "localhost:" + port
	err = go_test.WaitForTCPServer(serverHost, 3*time.Second)
	g.Expect(err).ToNot(HaveOccurred())
	testFun("http://" + serverHost)
}
