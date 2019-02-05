package main_test

import (
	"bytes"
	"fmt"
	"github.com/pivotal/indicator-protocol/pkg/registry"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal/indicator-protocol/pkg/go_test"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
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

		session := run(g, directory)
		defer session.Kill()

		g.Eventually(session).Should(gexec.Exit(0))

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(1))

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, files[0].Name()))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(string(data)).To(ContainSubstring(`expr: test_query{deployment="test_deployment"}`))
	})

	t.Run("reads and writes multiple documents to output directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		address := ":12345"
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

		session := run(g, directory)
		defer session.Kill()

		g.Eventually(session).Should(gexec.Exit(0))

		files, err := ioutil.ReadDir(directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(files).To(HaveLen(3))
	})

	t.Run("saves documents with expected file names", func(t *testing.T) {
		g := NewGomegaWithT(t)

		address := ":12345"
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

		session := run(g, directory)
		defer session.Kill()

		g.Eventually(session).Should(gexec.Exit(0))

		fileNames := getFileNames(g, directory)

		g.Expect(fileNames).To(ConsistOf("test_product_0.yml", "test_product_1.yml", "test_product_2.yml"))

	})

	t.Run("write correct rule", func(t *testing.T) {
		g := NewGomegaWithT(t)

		address := ":12345"
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

		session := run(g, directory)
		defer session.Kill()

		g.Eventually(session).Should(gexec.Exit(0))

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

	t.Run("writes correctly formatted comparator to correct file", func(t *testing.T){
		g := NewGomegaWithT(t)

		address := ":12345"
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

		session := run(g, directory)
		defer session.Kill()

		g.Eventually(session).Should(gexec.Exit(0))

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
		fmt.Println(string(data))
		g.Expect(string(data)).To(ContainSubstring("> 5"))
	})

}

func getFileNames(g *GomegaWithT, directory string) []string{
	files, err := ioutil.ReadDir(directory)
	g.Expect(err).ToNot(HaveOccurred())
	fileNames := make([]string, 0)
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames
}

func run(g *GomegaWithT, outputDirectory string) *gexec.Session {
	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())
	cmd := exec.Command(
		binPath,
		"--registry", "https://localhost:12345",
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
