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
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

var (
	serverCert = "../../test_fixtures/leaf.pem"
	serverKey  = "../../test_fixtures/leaf.key"
	rootCACert = "../../test_fixtures/root.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestPrometheusRulesControllerBinary(t *testing.T) {
	t.Run("reads documents from registry and outputs to output-directory", func(t *testing.T) {
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

		registryAddress := "localhost:13245"
		config := registry.WebServerConfig{
			Address:       registryAddress,
			ServerPEMPath: serverCert,
			ServerKeyPath: serverKey,
			RootCAPath:    rootCACert,
			DocumentStore: store,
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

		prometheusServer := ghttp.NewServer()
		defer prometheusServer.Close()

		prometheusServer.AppendHandlers(func(w http.ResponseWriter, r *http.Request) {
			g.Expect(r.Method).To(Equal("POST"))
			g.Expect(r.URL.Path).To(Equal("/-/reload"))

			w.WriteHeader(http.StatusOK)
		})

		directory, err := ioutil.TempDir("", "test-alerts")
		g.Expect(err).ToNot(HaveOccurred())

		session := run(g, directory, fmt.Sprintf("https://%s", registryAddress), prometheusServer.URL())
		defer session.Kill()

		err = go_test.WaitForFiles(directory, 1)
		g.Expect(err).ToNot(HaveOccurred())

		fs := osfs.New("/")
		files, err := fs.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))

		file, err := fs.Open(fmt.Sprintf("%s/test_product_5482021faf855c4956467ffa12adef6cd9c559b2.yml", directory))
		g.Expect(err).ToNot(HaveOccurred())
		data, err := ioutil.ReadAll(file)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(string(data)).To(MatchYAML(`groups:
                - name: test_product
                  rules:
                  - alert: test_indicator
                    expr: test_query{deployment="test_deployment"} < 5
                    labels:
                      level: critical
                      product: test_product
                      version: v1.2.3
                    annotations:
                      test1: a
                      test2: b`))

		g.Eventually(prometheusServer.ReceivedRequests, 5*time.Second, 50*time.Millisecond).Should(HaveLen(1))
	})
}

func run(g *GomegaWithT, outputDirectory, registryURL, prometheusURL string) *gexec.Session {
	binPath, err := go_test.Build("./", "-race")
	g.Expect(err).ToNot(HaveOccurred())
	cmd := exec.Command(
		binPath,
		"--registry", registryURL,
		"--prometheus", prometheusURL,
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
