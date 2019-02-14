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
	"github.com/pivotal/indicator-protocol/pkg/go_test"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/registry"
)

var (
	serverCert = "../../test_fixtures/leaf.pem"
	serverKey  = "../../test_fixtures/leaf.key"
	rootCACert = "../../test_fixtures/root.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestGrafanaDashboardControllerBinary(t *testing.T) {
	t.Run("reads documents from registry and outputs graph files to output-directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(5 * time.Second)

		store.UpsertDocument(indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    "test_product",
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: `test_query{deployment="test_deployment"}`,
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.LessThan,
					Value:    5,
				}},
				Documentation: map[string]string{
					"test1": "a",
					"test2": "b",
				},
			}},
		})

		registryAddress := "localhost:12346"
		config := registry.WebServerConfig{
			Address:       registryAddress,
			ServerPEMPath: serverCert,
			ServerKeyPath: serverKey,
			RootCAPath:    rootCACert,
			DocumentStore: store,
		}

		start, stop, err := registry.NewWebServer(config)
		g.Expect(err).ToNot(HaveOccurred())

		defer func() {_ = stop()}()
		go func() {
			err := start()
			g.Expect(err).ToNot(HaveOccurred())
		}()

		directory, err := ioutil.TempDir("", "test-dashboards")
		g.Expect(err).ToNot(HaveOccurred())

		session := run(g, fmt.Sprintf("https://%s", registryAddress), directory)
		defer session.Kill()

		err = go_test.WaitForFiles(directory, 1)
		g.Expect(err).ToNot(HaveOccurred())

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/test_product_f3eb510a2597b4a81945dd616fbce2dbcb941c5e.json", directory))
		g.Expect(err).ToNot(HaveOccurred())

		fileBytes, err := ioutil.ReadFile("test_fixtures/expected_dashboard.json")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(string(data)).To(MatchJSON(fileBytes))
	})
}

func run(g *GomegaWithT, registryURL, outputDirectory string) *gexec.Session {
	binPath, err := go_test.Build("./")
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
