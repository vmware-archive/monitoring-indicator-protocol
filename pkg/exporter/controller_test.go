package exporter_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/exporter"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestController(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	t.Run("Start runs on an interval until stopped", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1, "v1alpha1"),
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

		var rcCalls, mrCalls int
		g.Eventually(func() int {
			rcCalls = registryClient.calls()
			mrCalls = mockReloader.calls()
			return mrCalls
		}).Should(BeNumerically(">", 4))
		g.Expect(mrCalls).To(Equal(rcCalls))
	})

	t.Run("saves multiple documents with expected file names", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(3, "v1alpha1"),
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

	t.Run("saves v0 documents", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1, "v0"),
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
		g.Expect(fileNames).To(ConsistOf("test_product_0.yml"))
	})

	t.Run("Update removes outdated files from output directory", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: []registry.APIV0Document{{
				APIVersion: "v1alpha1",
				Product: registry.APIV0Product{
					Name:    "test_product_A",
					Version: "v1.2.3",
				},
				Indicators: []registry.APIV0Indicator{{
					Name:         "test_indicator",
					PromQL:       `test_query{deployment="test_deployment"}`,
					Alert:        test_fixtures.DefaultAPIV0Alert(),
					Presentation: test_fixtures.DefaultAPIV0Presentation(),
					Thresholds: []registry.APIV0Threshold{{
						Level:    "critical",
						Operator: "lt",
						Value:    5,
					}},
				}},
				Layout: test_fixtures.DefaultAPIV0Layout([]string{"test_indicator"}),
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
			APIVersion: "v1alpha1",
			Product: registry.APIV0Product{
				Name:    "test_product_B",
				Version: "v1.2.3",
			},
			Indicators: []registry.APIV0Indicator{{
				Name:         "test_indicator",
				PromQL:       `test_query{deployment="test_deployment"}`,
				Alert:        test_fixtures.DefaultAPIV0Alert(),
				Presentation: test_fixtures.DefaultAPIV0Presentation(),
				Thresholds: []registry.APIV0Threshold{{
					Level:    "critical",
					Operator: "lt",
					Value:    5,
				}},
			}},
			Layout: test_fixtures.DefaultAPIV0Layout([]string{"test_indicator"}),
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
				APIVersion: "v1alpha1",
				Product: registry.APIV0Product{
					Name:    "test_product_A",
					Version: "v1.2.3",
				},
				Indicators: []registry.APIV0Indicator{{
					Name:         "test_indicator",
					PromQL:       `test_query{deployment="test_deployment"}`,
					Alert:        test_fixtures.DefaultAPIV0Alert(),
					Presentation: test_fixtures.DefaultAPIV0Presentation(),
					Thresholds: []registry.APIV0Threshold{{
						Level:    "critical",
						Operator: "lt",
						Value:    5,
					}},
				}},
				Layout: test_fixtures.DefaultAPIV0Layout([]string{"test_indicator"}),
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
		g.Expect(err).ToNot(MatchError(ContainSubstring("registry error response test")))
		g.Expect(err.Error()).To(Equal("failed to fetch indicator documents"))

		fileNames, err = go_test.GetFileNames(fs, directory)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(fileNames).To(ConsistOf("test_product_A.yml"))
	})
}

var stubConverter = func(document v1alpha1.IndicatorDocument) (*exporter.File, error) {
	return &exporter.File{Name: fmt.Sprintf("%s.yml", document.Spec.Product.Name), Contents: []byte("")}, nil
}

func TestReloading(t *testing.T) {
	t.Run("reloads after updating", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1, "v1alpha1"),
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

		g.Expect(mockReloader.calls()).To(Equal(1))
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

		g.Expect(mockReloader.calls()).To(Equal(0))
	})

	t.Run("returns an error if reload fails", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryClient := &mockRegistryClient{
			Documents: createTestDocuments(1, "v1alpha1"),
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

		g.Expect(mockReloader.calls()).To(Equal(1))
	})
}

var testComparators = []string{"lt", "lte", "eq", "neq", "gte", "gt"}

func createTestDocuments(count int, apiVersion string) []registry.APIV0Document {
	docs := make([]registry.APIV0Document, count)
	for i := 0; i < count; i++ {
		indicatorName := fmt.Sprintf("test_indicator_%d", i)
		docs[i] = registry.APIV0Document{
			APIVersion: apiVersion,
			Product: registry.APIV0Product{
				Name:    fmt.Sprintf("test_product_%d", i),
				Version: "v1.2.3",
			},
			Metadata: map[string]string{"deployment": "test_deployment"},
			Indicators: []registry.APIV0Indicator{{
				Name:   indicatorName,
				PromQL: `test_query{deployment="test_deployment"}`,
				Alert:  test_fixtures.DefaultAPIV0Alert(),
				Thresholds: []registry.APIV0Threshold{{
					Level:    "critical",
					Operator: testComparators[i],
					Value:    5,
				}},
				Presentation: test_fixtures.DefaultAPIV0Presentation(),
				Documentation: map[string]string{
					"test1": "a",
					"test2": "b",
				},
			}},
			Layout: test_fixtures.DefaultAPIV0Layout([]string{indicatorName}),
		}
	}
	return docs
}

type mockRegistryClient struct {
	Documents []registry.APIV0Document
	Error     error

	mu     sync.Mutex
	calls_ int
}

func (a *mockRegistryClient) IndicatorDocuments() ([]registry.APIV0Document, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls_ = a.calls_ + 1
	return a.Documents, a.Error
}

func (a *mockRegistryClient) calls() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.calls_
}

type mockReloader struct {
	fail bool

	mu     sync.Mutex
	calls_ int
}

func (a *mockReloader) Reload() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls_ = a.calls_ + 1

	if a.fail {
		return errors.New("")
	}

	return nil
}

func (a *mockReloader) calls() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.calls_
}
