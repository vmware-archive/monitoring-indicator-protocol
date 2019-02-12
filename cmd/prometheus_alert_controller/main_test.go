package main_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
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

func TestPrometheusAlertControllerBinary(t *testing.T) {
	t.Run("reads documents from registry and outputs to output-directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		address := ":12345"
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

		config := registry.WebServerConfig{
			Address:       address,
			ServerPEMPath: serverCert,
			ServerKeyPath: serverKey,
			RootCAPath:    rootCACert,
			DocumentStore: store,
		}

		start, stop, err := registry.NewWebServer(config)
		g.Expect(err).ToNot(HaveOccurred())

		defer stop()
		go start()

		prometheusServer := ghttp.NewServer()
		defer prometheusServer.Close()

		prometheusServer.AppendHandlers(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			g.Expect(r.Method).To(Equal("POST"))
			g.Expect(r.URL.Path).To(Equal("/-/reload"))

			w.WriteHeader(http.StatusOK)
		})

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		session := run(g, directory, prometheusServer.URL())
		defer session.Kill()

		time.Sleep(50 * time.Millisecond)

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, files[0].Name()))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(string(data)).To(ContainSubstring(`expr: test_query{deployment="test_deployment"}`))

		g.Expect(prometheusServer.ReceivedRequests()).To(HaveLen(1))
	})
}

func run(g *GomegaWithT, outputDirectory string, prometheusURI string) *gexec.Session {
	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())
	cmd := exec.Command(
		binPath,
		"--registry", "https://localhost:12345",
		"--prometheus", prometheusURI,
		"--tls-pem-path", clientCert,
		"--tls-key-path", clientKey,
		"--tls-root-ca-pem", rootCACert,
		"--tls-server-cn", "localhost",
		"--output-directory", outputDirectory,
	)

	buffer := bytes.NewBuffer(nil)
	session, err := gexec.Start(cmd, buffer, buffer)

	g.Expect(err).ToNot(HaveOccurred())
	return session
}
