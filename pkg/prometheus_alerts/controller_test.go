package prometheus_alerts_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal/indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/indicator-protocol/pkg/registry"
)

func TestAlertController(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	t.Run("reads and writes multiple documents to output directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(3),
		}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(3))
	})

	t.Run("saves documents with expected file names", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(3),
		}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		fileNames := getFileNames(g, directory)

		g.Expect(fileNames).To(ConsistOf("test_product_0.yml", "test_product_1.yml", "test_product_2.yml"))

	})

	t.Run("write correct rule", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1),
		}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "test_product_0.yml"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(
			ContainSubstring(
				`
  rules:
  - alert: test_indicator_0
    expr: test_query{deployment="test_deployment"} < 5
    labels:
      level: critical
      product: test_product_0
      version: v1.2.3
    annotations:
      test1: a
      test2: b`))
	})

	t.Run("writes correctly formatted comparator to correct file", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(6),
		}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "test_product_0.yml"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(ContainSubstring("< 5"))

		data, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "test_product_1.yml"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(ContainSubstring("<= 5"))

		data, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "test_product_2.yml"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(ContainSubstring("== 5"))

		data, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "test_product_3.yml"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(ContainSubstring("!= 5"))

		data, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "test_product_4.yml"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(ContainSubstring(">= 5"))

		data, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, "test_product_5.yml"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(data)).To(ContainSubstring("> 5"))
	})

	t.Run("Update removes outdated files from output directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		initialRegistryStateClient := &mockRegistryClient{
			Documents: []registry.APIV0Document{{
				APIVersion: "v0",
				Product: registry.APIV0Product{
					Name:    "test_product_A",
					Version: "v1.2.3",
				},
				Indicators: []registry.APIV0Indicator{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Thresholds: []registry.APIV0Threshold{{
						Level:    "critical",
						Operator: "lt",
						Value:    5,
					}},
				}},
			}},
		}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   initialRegistryStateClient,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))
		g.Expect(files[0].Name()).To(Equal("test_product_A.yml"))

		newRegistryStateClient := &mockRegistryClient{
			Documents: []registry.APIV0Document{{
				APIVersion: "v0",
				Product: registry.APIV0Product{
					Name:    "test_product_B",
					Version: "v1.2.3",
				},
				Indicators: []registry.APIV0Indicator{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Thresholds: []registry.APIV0Threshold{{
						Level:    "critical",
						Operator: "lt",
						Value:    5,
					}},
				}},
			}},
		}

		c = prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   newRegistryStateClient,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller = prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		files, err = ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))
		g.Expect(files[0].Name()).To(Equal("test_product_B.yml"))
	})

	t.Run("leaves documents in output directory if registry returns error", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: []registry.APIV0Document{{
				APIVersion: "v0",
				Product: registry.APIV0Product{
					Name:    "test_product_A",
					Version: "v1.2.3",
				},
				Indicators: []registry.APIV0Indicator{{
					Name:   "test_indicator",
					PromQL: `test_query{deployment="test_deployment"}`,
					Thresholds: []registry.APIV0Threshold{{
						Level:    "critical",
						Operator: "lt",
						Value:    5,
					}},
				}},
			}},
		}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))
		g.Expect(files[0].Name()).To(Equal("test_product_A.yml"))

		errorRegistry := &mockRegistryClient{
			Error: fmt.Errorf("registry error response test"),
		}

		c = prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   errorRegistry,
			PrometheusAPIClient: &mockPrometheusClient{},
			OutputDirectory:     directory,
		}

		controller = prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).To(MatchError(ContainSubstring("registry error response test")))

		files, err = ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))
		g.Expect(files[0].Name()).To(Equal("test_product_A.yml"))
	})
}

func TestPrometheusClientIntegration(t *testing.T) {
	t.Run("it POSTs to prometheus reload endpoint when done", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1),
		}

		prometheusClient := &mockPrometheusClient{}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: prometheusClient,
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(prometheusClient.Calls).To(Equal(1))
	})

	t.Run("does not post to reload if there is an error getting documents", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: nil,
			Error:     fmt.Errorf("oh no! this is bad"),
		}

		prometheusClient := &mockPrometheusClient{}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: prometheusClient,
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)

		err = controller.Update()
		g.Expect(err).To(HaveOccurred())

		g.Expect(prometheusClient.Calls).To(Equal(0))
	})

	t.Run("returns an error if Prometheus reload fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1),
		}

		prometheusClient := &mockPrometheusClient{
			Error: fmt.Errorf("oh no! this is bad, too bad"),
		}

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryAPIClient:   registryClient,
			PrometheusAPIClient: prometheusClient,
			OutputDirectory:     directory,
		}

		controller := prometheus_alerts.NewController(c)

		err = controller.Update()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(ContainSubstring("oh no! this is bad, too bad")))

		g.Expect(prometheusClient.Calls).To(Equal(1))
	})
}

func TestPrometheusClient(t *testing.T) {
	t.Run("it posts to /-/reload", func(t *testing.T) {
		g := NewGomegaWithT(t)

		prometheusServer := ghttp.NewServer()
		defer prometheusServer.Close()

		prometheusServer.AppendHandlers(fakePrometheusReload(g, http.StatusOK))

		client := prometheus_alerts.NewPrometheusClient(prometheusServer.URL(), &http.Client{})
		err := client.Reload()
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(prometheusServer.ReceivedRequests()).To(HaveLen(1))
	})

	t.Run("it returns an error if prometheus responds with an error", func(t *testing.T) {
		g := NewGomegaWithT(t)

		prometheusServer := ghttp.NewServer()
		defer prometheusServer.Close()

		prometheusServer.AppendHandlers(fakePrometheusReload(g, http.StatusInternalServerError))

		client := prometheus_alerts.NewPrometheusClient(prometheusServer.URL(), &http.Client{})
		err := client.Reload()
		g.Expect(err).To(HaveOccurred())
	})
}

func fakePrometheusReload(g *GomegaWithT, response int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := r.Body.Close()
			g.Expect(err).ToNot(HaveOccurred())
		}()
		g.Expect(r.Method).To(Equal("POST"))
		g.Expect(r.URL.Path).To(Equal("/-/reload"))

		w.WriteHeader(response)
	}
}

func getFileNames(g *GomegaWithT, directory string) []string {
	files, err := ioutil.ReadDir(directory)
	g.Expect(err).ToNot(HaveOccurred())
	fileNames := make([]string, 0)
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames
}

var testComparators = []string{"lt", "lte", "eq", "neq", "gte", "gt"}

func createTestDocuments(count int) []registry.APIV0Document {
	docs := make([]registry.APIV0Document, count)
	for i := 0; i < count; i++ {
		docs[i] = registry.APIV0Document{
			APIVersion: "v0",
			Product: registry.APIV0Product{
				Name:    fmt.Sprintf("test_product_%d", i),
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []registry.APIV0Indicator{{
				Name:   fmt.Sprintf("test_indicator_%d", i),
				PromQL: `test_query{deployment="test_deployment"}`,
				Thresholds: []registry.APIV0Threshold{{
					Level:    "critical",
					Operator: testComparators[i],
					Value:    5,
				}},
				Documentation: map[string]string{
					"test1": "a",
					"test2": "b",
				},
			}},
		}
	}
	return docs
}

type mockRegistryClient struct {
	Documents []registry.APIV0Document
	Error     error
}

func (a mockRegistryClient) IndicatorDocuments() ([]registry.APIV0Document, error) {
	return a.Documents, a.Error
}

type mockPrometheusClient struct {
	Error error
	Calls int
}

func (p *mockPrometheusClient) Reload() error {
	p.Calls = p.Calls + 1
	return p.Error
}
