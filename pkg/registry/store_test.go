package registry_test

import (
	"testing"
	"time"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func TestStore(t *testing.T) {
	var val interface{}
	val = map[interface{}]interface{}{
		"promQL": "foo{bar&bar}",
		"documentation": map[interface{}]interface{}{
			"title": "Great Success",
		},
	}

	productName := "test-app"
	productVersion := "test-version"
	patchA := indicator.Patch{
		Match: indicator.Match{
			Name:    &productName,
			Version: &productVersion,
		},
		Operations: []patch.OpDefinition{{
			Type:  "replace",
			Path:  strPtr("indicators/name=success_percentage"),
			Value: &val,
		}},
	}

	patchB := indicator.Patch{
		Match: indicator.Match{
			Name:    &productName,
			Version: &productVersion,
		},
		Operations: []patch.OpDefinition{{
			Type:  "replace",
			Path:  strPtr("indicators/name=success_percentage"),
			Value: &val,
		}},
	}

	var newIndicator interface{}
	newIndicator = map[interface{}]interface{}{
		"name":   "another_indicator",
		"promQL": "foo{bar&bar}",
	}
	patchC := indicator.Patch{
		Match: indicator.Match{
			Name:    &productName,
			Version: &productVersion,
		},
		Operations: []patch.OpDefinition{{
			Type:  "add",
			Path:  strPtr("indicators/-"),
			Value: &newIndicator,
		}},
	}

	productAVersion1Document := v1alpha1.IndicatorDocument{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"deployment": "abc-123",
			},
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{Name: "my-product-a", Version: "1"},
			Indicators: []v1alpha1.IndicatorSpec{{
				Name: "test_errors",
			}},
		},
	}

	productAVersion2Document := v1alpha1.IndicatorDocument{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"deployment": "abc-123",
			},
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{Name: "my-product-a", Version: "2"},
			Indicators: []v1alpha1.IndicatorSpec{{
				Name: "test_error_ratio",
			}},
		},
	}

	productADeployment2Document := v1alpha1.IndicatorDocument{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"deployment": "def-456",
			},
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{Name: "my-product-a", Version: "2"},
			Indicators: []v1alpha1.IndicatorSpec{{
				Name: "test_error_ratio",
			}},
		},
	}

	productBDocument := v1alpha1.IndicatorDocument{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"deployment": "def-456",
			},
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{Name: "my-product-b", Version: "1"},
			Indicators: []v1alpha1.IndicatorSpec{{
				Name: "test_latency",
			}},
		},
	}

	t.Run("it upserts patchesBySource in bulk by source", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(time.Hour, time.Now)

		store.UpsertPatches(registry.PatchList{
			Source:  "git:other-repo",
			Patches: []indicator.Patch{patchB, patchC},
		})
		g.Expect(store.AllPatches()).To(ConsistOf(patchB, patchC))

		store.UpsertPatches(registry.PatchList{
			Source:  "git:other-repo",
			Patches: []indicator.Patch{patchB},
		})
		g.Expect(store.AllPatches()).To(ConsistOf(patchB))

		store.UpsertPatches(registry.PatchList{
			Source:  "git:repo",
			Patches: []indicator.Patch{patchA},
		})
		g.Expect(store.AllPatches()).To(ConsistOf(patchB, patchA))
	})

	t.Run("it saves documents sent to it", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(time.Hour, time.Now)

		store.UpsertDocument(productAVersion1Document)

		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document))
	})

	t.Run("it can retrieve documents by product name", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := registry.NewDocumentStore(time.Hour, time.Now)

		store.UpsertDocument(productAVersion1Document)
		store.UpsertDocument(productBDocument)

		g.Expect(store.FilteredDocuments("my-product-a")).To(ConsistOf(productAVersion1Document))
	})

	t.Run("it upserts documents based on product", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(time.Hour, time.Now)

		store.UpsertDocument(productAVersion1Document)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document))

		store.UpsertDocument(productBDocument)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document, productBDocument))

		store.UpsertDocument(productAVersion2Document)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion2Document, productBDocument))
	})

	t.Run("it upserts documents based on metadata", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(time.Hour, time.Now)

		store.UpsertDocument(productAVersion1Document)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document))

		store.UpsertDocument(productADeployment2Document)
		g.Expect(store.AllDocuments()).To(ConsistOf(productAVersion1Document, productADeployment2Document))
	})

	t.Run("documents expire after an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		theTime := time.Now()
		store := registry.NewDocumentStore(time.Hour, func() time.Time { return theTime })

		store.UpsertDocument(productAVersion1Document)
		theTime = theTime.Add(time.Hour).Add(time.Millisecond)

		g.Expect(store.AllDocuments()).To(HaveLen(0))
	})
}

func strPtr(s string) *string {
	return &s
}
