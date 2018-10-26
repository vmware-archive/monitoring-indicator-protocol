package registry_test

import (
	. "github.com/onsi/gomega"
	"testing"
	"time"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/registry"
)

func TestInsertDocument(t *testing.T) {

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

		store.Upsert(productAVersion1Document)

		g.Expect(store.All()).To(ConsistOf(productAVersion1Document))
	})

	t.Run("it upserts documents based on labels", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(10 * time.Millisecond)

		store.Upsert(productAVersion1Document)
		g.Expect(store.All()).To(ConsistOf(productAVersion1Document))

		store.Upsert(productBDocument)
		g.Expect(store.All()).To(ConsistOf(productAVersion1Document, productBDocument))

		store.Upsert(productAVersion2Document)
		g.Expect(store.All()).To(ConsistOf(productAVersion2Document, productBDocument))
	})

	t.Run("documents expire after an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := registry.NewDocumentStore(10 * time.Millisecond)

		store.Upsert(productAVersion1Document)
		time.Sleep(11 * time.Millisecond)

		g.Expect(store.All()).To(HaveLen(0))
	})
}
