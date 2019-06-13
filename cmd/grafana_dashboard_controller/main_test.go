package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

var (
	serverCert = "../../test_fixtures/server.pem"
	serverKey  = "../../test_fixtures/server.key"
	rootCACert = "../../test_fixtures/ca.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestGrafanaDashboardControllerBinary(t *testing.T) {
	t.Run("reads documents from registry and outputs graph files to output-directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(time.Hour, time.Now)

		document := indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "test_product",
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: `test_query{deployment="test_deployment"}`,
				Alert:  test_fixtures.DefaultAlert(),
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.LessThan,
					Value:    5,
				}},
				Presentation:  test_fixtures.DefaultPresentation(),
				Documentation: map[string]string{"title": "Test Indicator Title"},
			}},
			Layout: indicator.Layout{
				Title: "Test Dashboard",
				Sections: []indicator.Section{
					{
						Title:      "Test Section Title",
						Indicators: []string{"test_indicator"},
					},
				},
			},
		}

		store.UpsertDocument(document)

		registryAddress := "localhost:12346"
		config := registry.WebServerConfig{
			Address:       registryAddress,
			DocumentStore: store,
			StatusStore:   status_store.New(time.Now),
		}

		start, stop, err := registry.NewWebServer(config)
		g.Expect(err).ToNot(HaveOccurred())

		done := make(chan struct{})
		defer func() {
			_ = stop()
			<-done
		}()
		go func() {
			defer close(done)
			_ = start()
		}()

		directory, err := ioutil.TempDir("", "test-dashboards")
		g.Expect(err).ToNot(HaveOccurred())

		session := run(g, fmt.Sprintf("http://%s", registryAddress), directory)
		defer session.Kill()

		err = go_test.WaitForFiles(directory, 1)
		g.Expect(err).ToNot(HaveOccurred())

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, files[0].Name()))
		g.Expect(err).ToNot(HaveOccurred())

		fileBytes, err := ioutil.ReadFile("test_fixtures/expected_dashboard.json")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(string(data)).To(MatchJSON(fileBytes))
	})
}

func run(g *GomegaWithT, registryURL, outputDirectory string) *gexec.Session {
	binPath, err := go_test.Build("./", "-race")
	g.Expect(err).ToNot(HaveOccurred())
	cmd := exec.Command(
		binPath,
		"--registry", registryURL,
		"--tls-pem-path", clientCert,
		"--tls-key-path", clientKey,
		"--tls-root-ca-pem", rootCACert,
		"--tls-server-cn", "localhost",
		"--output-directory", outputDirectory,
	)

	session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

	g.Expect(err).ToNot(HaveOccurred())
	return session
}
