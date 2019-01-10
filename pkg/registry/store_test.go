package registry_test

import (
	. "github.com/onsi/gomega"
	"testing"
	"time"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/registry"
	"github.com/krishicks/yaml-patch"
)

func TestInsertDocument(t *testing.T) {
	var val interface{}
	val = indicator.Indicator{
		PromQL: "foo{bar&bar}",
		Documentation: map[string]string{
			"title": "Great Success",
		},
	}

	productName := "test-app"
	productVersion := "test-version"
	patchAVer1 := indicator.Patch{
		Origin: "git:repo/file.yml",
		Match: indicator.Match{
			Name:    &productName,
			Version: &productVersion,
		},
		Operations: []yamlpatch.Operation{{
			Op:    "replace",
			Path:  "indicators/name=success_percentage",
			Value: yamlpatch.NewNode(&val),
		}},
	}

	patchAVer2 := indicator.Patch{
		Origin: "git:repo/file.yml",
		Match: indicator.Match{
			Name:    &productName,
			Version: &productVersion,
		},
		Operations: []yamlpatch.Operation{{
			Op:    "replace",
			Path:  "indicators/name=succsoops_percentage",
			Value: yamlpatch.NewNode(&val),
		}},
	}

	patchB := indicator.Patch{
		Origin: "git:other-repo/file.yml",
		Match: indicator.Match{
			Name:    &productName,
			Version: &productVersion,
		},
		Operations: []yamlpatch.Operation{{
			Op:    "replace",
			Path:  "indicators/name=success_percentage",
			Value: yamlpatch.NewNode(&val),
		}},
	}

	productAVersion1Document := indicator.Document{
		Product: indicator.Product{Name: "my-product-a", Version: "1"},
		Metadata: map[string]string{
			"deployment": "abc-123",
		},
		Indicators: []indicator.Indicator{{
			Name: "test_errors",
		}},
	}

	productAVersion2Document := indicator.Document{
		Product: indicator.Product{Name: "my-product-a", Version: "2"},
		Metadata: map[string]string{
			"deployment": "abc-123",
		},
		Indicators: []indicator.Indicator{{
			Name: "test_error_ratio",
		}},
	}

	productBDocument := indicator.Document{
		Product: indicator.Product{Name: "my-product-b", Version: "1"},
		Metadata: map[string]string{
			"deployment": "def-456",
		},
		Indicators: []indicator.Indicator{{
			Name: "test_latency",
		}},
	}

	t.Run("it saves documents sent to it", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(10 * time.Millisecond)

		store.UpsertDocument(productAVersion1Document)

		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document))
	})

	t.Run("it upserts documents based on labels", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(10 * time.Millisecond)

		store.UpsertDocument(productAVersion1Document)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document))

		store.UpsertDocument(productBDocument)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document, productBDocument))

		store.UpsertDocument(productAVersion2Document)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion2Document, productBDocument))
	})

	t.Run("documents expire after an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(10 * time.Millisecond)

		store.UpsertDocument(productAVersion1Document)
		time.Sleep(11 * time.Millisecond)

		g.Expect(store.AllDocuments()).To(HaveLen(0))
	})

	t.Run("it saves inserted patches", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(10 * time.Millisecond)

		store.UpsertPatch(patchAVer1)

		g.Expect(store.AllPatches()).To(ConsistOf(patchAVer1))
	})

	t.Run("it upserts patches based on origin", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(10 * time.Millisecond)

		store.UpsertPatch(patchAVer1)
		g.Expect(store.AllPatches()).To(ConsistOf(patchAVer1))

		store.UpsertPatch(patchB)
		g.Expect(store.AllPatches()).To(ConsistOf(patchAVer1, patchB))

		store.UpsertPatch(patchAVer2)
		g.Expect(store.AllPatches()).To(ConsistOf(patchAVer2, patchB))
	})
}
