package prometheus_alerts_test

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/indicator-protocol/pkg/registry"
)

var (
	serverCert = "../../test_fixtures/leaf.pem"
	serverKey  = "../../test_fixtures/leaf.key"
	rootCACert = "../../test_fixtures/root.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestAlertController(t *testing.T) {
	address := "localhost:12347"
	t.Run("reads and writes multiple documents to output directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(5 * time.Second)

		upsertDocuments(store, 3)

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

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryURI:       fmt.Sprintf("https://%s", address),
			TLSPEMPath:        clientCert,
			TLSKeyPath:        clientKey,
			TLSRootCACertPath: rootCACert,
			TLSServerCN:       "localhost",
			OutputDirectory:   directory,
		}

		controller := prometheus_alerts.NewController(c)
		controller.Update()

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(3))
	})

	t.Run("saves documents with expected file names", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(5 * time.Second)

		count := 3
		upsertDocuments(store, count)

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

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryURI:       fmt.Sprintf("https://%s", address),
			TLSPEMPath:        clientCert,
			TLSKeyPath:        clientKey,
			TLSRootCACertPath: rootCACert,
			TLSServerCN:       "localhost",
			OutputDirectory:   directory,
		}

		controller := prometheus_alerts.NewController(c)
		controller.Update()

		fileNames := getFileNames(g, directory)

		g.Expect(fileNames).To(ConsistOf("test_product_0.yml", "test_product_1.yml", "test_product_2.yml"))

	})

	t.Run("write correct rule", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(5 * time.Second)

		upsertDocuments(store, 1)

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

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryURI:       fmt.Sprintf("https://%s", address),
			TLSPEMPath:        clientCert,
			TLSKeyPath:        clientKey,
			TLSRootCACertPath: rootCACert,
			TLSServerCN:       "localhost",
			OutputDirectory:   directory,
		}

		controller := prometheus_alerts.NewController(c)
		controller.Update()

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

		store := registry.NewDocumentStore(5 * time.Second)

		count := 6
		upsertDocuments(store, count)

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

		directory, err := ioutil.TempDir("", "test")
		g.Expect(err).ToNot(HaveOccurred())

		c := prometheus_alerts.ControllerConfig{
			RegistryURI:       fmt.Sprintf("https://%s", address),
			TLSPEMPath:        clientCert,
			TLSKeyPath:        clientKey,
			TLSRootCACertPath: rootCACert,
			TLSServerCN:       "localhost",
			OutputDirectory:   directory,
		}

		controller := prometheus_alerts.NewController(c)
		controller.Update()

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

func upsertDocuments(store *registry.DocumentStore, count int) {
	for i := 0; i < count; i++ {
		store.UpsertDocument(indicator.Document{
			APIVersion: "v0",
			Product: indicator.Product{
				Name:    fmt.Sprintf("test_product_%d", i),
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{{
				Name:   fmt.Sprintf("test_indicator_%d", i),
				PromQL: `test_query{deployment="test_deployment"}`,
				Thresholds: []indicator.Threshold{{
					Level:    "critical",
					Operator: indicator.OperatorType(i),
					Value:    5,
				}},
				Documentation: map[string]string{
					"test1": "a",
					"test2": "b",
				},
			}},
		})
	}
}
