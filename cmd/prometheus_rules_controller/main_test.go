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
	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"gopkg.in/src-d/go-billy.v4/osfs"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

var (
	rootCACert = "../../test_fixtures/ca.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestPrometheusRulesControllerBinary(t *testing.T) {
	t.Run("reads documents from registry and outputs to output-directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(time.Hour, time.Now)

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: api_versions.V1alpha1,
				Kind:       "IndicatorDocument",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{

				Product: v1alpha1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"[$step]}`,
					Thresholds: []v1alpha1.Threshold{{
						Level:    "critical",
						Operator: v1alpha1.LessThan,
						Value:    5,
					}},
					Alert: v1alpha1.Alert{
						For:  "10m",
						Step: "5m",
					},
					Documentation: map[string]string{
						"test1": "a",
						"test2": "b",
					},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
			},
		}
		doc.Spec.Layout = test_fixtures.DefaultLayout(doc.Spec.Indicators)
		store.UpsertDocument(doc)

		registryAddress := "localhost:13245"
		config := registry.WebServerConfig{
			Address:       registryAddress,
			DocumentStore: store,
			StatusStore:   status_store.New(time.Now),
		}

		start, stop := registry.NewWebServer(config)

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

		session := run(g, directory, fmt.Sprintf("http://%s", registryAddress), prometheusServer.URL())
		defer session.Kill()

		err = go_test.WaitForFiles(directory, 1)
		g.Expect(err).ToNot(HaveOccurred())

		fs := osfs.New("/")
		files, err := fs.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())

		file, err := fs.Open(fmt.Sprintf("%s/%s", directory, files[0].Name()))
		g.Expect(err).ToNot(HaveOccurred())
		data, err := ioutil.ReadAll(file)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(string(data)).To(MatchYAML(`groups:
                - name: test_product
                  rules:
                  - alert: test_indicator
                    expr: test_query{deployment="test_deployment"[5m]} < 5
                    for: 10m
                    labels:
                      level: critical
                      product: test_product
                      version: v1.2.3
                      deployment: test_deployment
                    annotations:
                      test1: a
                      test2: b`))

		g.Eventually(prometheusServer.ReceivedRequests, 5*time.Second, 50*time.Millisecond).Should(HaveLen(1))
	})

	t.Run("reads v0 documents", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(time.Hour, time.Now)

		doc := v1alpha1.IndicatorDocument{
			TypeMeta: v1.TypeMeta{
				APIVersion: api_versions.V0,
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{"deployment": "test_deployment"},
			},
			Spec: v1alpha1.IndicatorDocumentSpec{

				Product: v1alpha1.Product{
					Name:    "test_product",
					Version: "v1.2.3",
				},
				Indicators: []v1alpha1.IndicatorSpec{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"[$step]}`,
					Thresholds: []v1alpha1.Threshold{{
						Level:    "critical",
						Operator: v1alpha1.LessThan,
						Value:    5,
					}},
					Alert: v1alpha1.Alert{
						For:  "10m",
						Step: "5m",
					},
					Documentation: map[string]string{
						"test1": "a",
						"test2": "b",
					},
					Presentation: test_fixtures.DefaultPresentation(),
				}},
			},
		}
		doc.Spec.Layout = test_fixtures.DefaultLayout(doc.Spec.Indicators)
		store.UpsertDocument(doc)

		registryAddress := "localhost:13245"
		config := registry.WebServerConfig{
			Address:       registryAddress,
			DocumentStore: store,
			StatusStore:   status_store.New(time.Now),
		}

		start, stop := registry.NewWebServer(config)

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

		session := run(g, directory, fmt.Sprintf("http://%s", registryAddress), prometheusServer.URL())
		defer session.Kill()

		err = go_test.WaitForFiles(directory, 1)
		g.Expect(err).ToNot(HaveOccurred())

		fs := osfs.New("/")
		files, err := fs.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())

		file, err := fs.Open(fmt.Sprintf("%s/%s", directory, files[0].Name()))
		g.Expect(err).ToNot(HaveOccurred())
		data, err := ioutil.ReadAll(file)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(string(data)).To(MatchYAML(`groups:
                - name: test_product
                  rules:
                  - alert: test_indicator
                    expr: test_query{deployment="test_deployment"[5m]} < 5
                    for: 10m
                    labels:
                      level: critical
                      product: test_product
                      version: v1.2.3
                      deployment: test_deployment
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
