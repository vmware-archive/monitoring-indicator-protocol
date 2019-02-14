package exporter_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pivotal/indicator-protocol/pkg/exporter"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"log"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pivotal/indicator-protocol/pkg/go_test"
	"github.com/pivotal/indicator-protocol/pkg/registry"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func TestController(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	t.Run("Start runs on an interval until stopped", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1),
		}

		mockReloader := &mockReloader{}

		c := exporter.ControllerConfig{
			RegistryAPIClient: registryClient,
			Filesystem:        memfs.New(),
			OutputDirectory:   "/",
			UpdateFrequency:   50 * time.Millisecond,
			DocType:           "",
			Converter:         stubConverter,
			Reloader:          mockReloader.Reload,
		}

		controller := exporter.NewController(c)
		go controller.Start()

		time.Sleep(5 * time.Millisecond)
		g.Expect(registryClient.Calls).To(Equal(1))
		g.Expect(mockReloader.Calls).To(Equal(1))

		time.Sleep(50 * time.Millisecond)
		g.Expect(registryClient.Calls).To(Equal(2))
		g.Expect(mockReloader.Calls).To(Equal(2))

		time.Sleep(50 * time.Millisecond)
		g.Expect(registryClient.Calls).To(Equal(3))
		g.Expect(mockReloader.Calls).To(Equal(3))

		time.Sleep(50 * time.Millisecond)
		g.Expect(registryClient.Calls).To(Equal(4))
		g.Expect(mockReloader.Calls).To(Equal(4))
	})

	t.Run("saves multiple documents with expected file names", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(3),
		}

		fs := memfs.New()
		directory := "/test"
		err := fs.MkdirAll(directory, 0644)
		g.Expect(err).ToNot(HaveOccurred())

		mockReloader := &mockReloader{}

		c := exporter.ControllerConfig{
			RegistryAPIClient: registryClient,
			Filesystem:        fs,
			OutputDirectory:   directory,
			UpdateFrequency:   0,
			DocType:           "",
			Converter:         stubConverter,
			Reloader:          mockReloader.Reload,
		}

		controller := exporter.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		fileNames, err := go_test.GetFileNames(fs, directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileNames).To(ConsistOf("test_product_0.yml", "test_product_1.yml", "test_product_2.yml"))

	})

	t.Run("Update removes outdated files from output directory", func(t *testing.T) {
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

		fs := memfs.New()
		directory := "/test"
		err := fs.MkdirAll(directory, 0644)
		g.Expect(err).ToNot(HaveOccurred())

		c := exporter.ControllerConfig{
			RegistryAPIClient: registryClient,
			Filesystem:        fs,
			OutputDirectory:   directory,
			UpdateFrequency:   0,
			DocType:           "",
			Converter:         stubConverter,
		}

		controller := exporter.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		fileNames, err := go_test.GetFileNames(fs, directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileNames).To(ConsistOf("test_product_A.yml"))

		registryClient.Documents = []registry.APIV0Document{{
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
		}}

		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		fileNames, err = go_test.GetFileNames(fs, directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileNames).To(ConsistOf("test_product_B.yml"))
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

		fs := memfs.New()
		directory := "/test"
		err := fs.MkdirAll(directory, 0644)
		g.Expect(err).ToNot(HaveOccurred())

		c := exporter.ControllerConfig{
			RegistryAPIClient: registryClient,
			Filesystem:        fs,
			OutputDirectory:   directory,
			Converter:         stubConverter,
		}

		controller := exporter.NewController(c)
		err = controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		fileNames, err := go_test.GetFileNames(fs, directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileNames).To(ConsistOf("test_product_A.yml"))

		registryClient.Documents = nil
		registryClient.Error = fmt.Errorf("registry error response test")

		err = controller.Update()
		g.Expect(err).To(MatchError(ContainSubstring("registry error response test")))

		fileNames, err = go_test.GetFileNames(fs, directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileNames).To(ConsistOf("test_product_A.yml"))
	})
}

var stubConverter = func(document indicator.Document) (*exporter.File, error) {
	return &exporter.File{Name: fmt.Sprintf("%s.yml", document.Product.Name), Contents: []byte("")}, nil
}

func TestReloading(t *testing.T) {
	t.Run("reloads after updating", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1),
		}

		mockReloader := mockReloader{}

		c := exporter.ControllerConfig{
			RegistryAPIClient: registryClient,
			Filesystem:        memfs.New(),
			OutputDirectory:   "/",
			Reloader:          mockReloader.Reload,
			Converter:         stubConverter,
		}

		controller := exporter.NewController(c)
		err := controller.Update()
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(mockReloader.Calls).To(Equal(1))
	})

	t.Run("does not reload if there is an error getting documents", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: nil,
			Error:     fmt.Errorf("oh no! this is bad"),
		}

		mockReloader := &mockReloader{}

		c := exporter.ControllerConfig{
			RegistryAPIClient: registryClient,
			Filesystem:        memfs.New(),
			OutputDirectory:   "/",
			Reloader:          mockReloader.Reload,
		}

		controller := exporter.NewController(c)

		err := controller.Update()
		g.Expect(err).To(HaveOccurred())

		g.Expect(mockReloader.Calls).To(Equal(0))
	})

	t.Run("returns an error if reload fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1),
		}

		mockReloader := &mockReloader{
			fail: true,
		}

		c := exporter.ControllerConfig{
			RegistryAPIClient: registryClient,
			Filesystem:        memfs.New(),
			OutputDirectory:   "/",
			UpdateFrequency:   0,
			DocType:           "",
			Converter:         stubConverter,
			Reloader:          mockReloader.Reload,
		}

		controller := exporter.NewController(c)

		err := controller.Update()
		g.Expect(err).To(HaveOccurred())

		g.Expect(mockReloader.Calls).To(Equal(1))
	})
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
	Calls     int
}

func (a *mockRegistryClient) IndicatorDocuments() ([]registry.APIV0Document, error) {
	a.Calls = a.Calls + 1
	return a.Documents, a.Error
}

type mockReloader struct {
	Calls int
	fail  bool
}

func (a *mockReloader) Reload() error {
	a.Calls = a.Calls + 1

	if a.fail {
		return errors.New("")
	}

	return nil
}
