package registry_test

import (
	. "github.com/onsi/gomega"
	"testing"
	"time"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/registry"
)

func TestInsertDocument(t *testing.T) {
	d := registry.NewDocumentStore(10 * time.Millisecond)

	t.Run("it saves documents sent to it", func(t *testing.T) {
		g := NewGomegaWithT(t)

		d.Upsert(
			map[string]string{"test_label": "test_value"},
			[]indicator.Indicator{{
				Name:  "test_name",
				Title: "test_title",
			}},
		)

		docs := d.All()
		g.Expect(docs).To(HaveLen(1))
		g.Expect(docs[0].Indicators).To(Equal([]indicator.Indicator{{
			Name:  "test_name",
			Title: "test_title",
		}}))
		g.Expect(docs[0].Labels).To(Equal(map[string]string{"test_label": "test_value"}))
	})

	t.Run("it upserts documents based on labels", func(t *testing.T) {
		g := NewGomegaWithT(t)

		d.Upsert(
			map[string]string{"deployment": "cf-abc-123", "product": "pas"},
			[]indicator.Indicator{{
				Name: "test_name",
			}},
		)
		d.Upsert(
			map[string]string{"deployment": "cf-abc-123", "product": "pas"},
			[]indicator.Indicator{{
				Name: "router_latency",
			}, {
				Name: "diego_capacity",
			}},
		)

		docs := d.All()
		g.Expect(docs).To(HaveLen(2))
		g.Expect(docs[0].Indicators).To(Equal([]indicator.Indicator{{
			Name:  "test_name",
			Title: "test_title",
		}}))
		g.Expect(docs[0].Labels).To(Equal(map[string]string{"test_label": "test_value"}))
		g.Expect(docs[1].Indicators).To(Equal([]indicator.Indicator{{
			Name: "router_latency",
		}, {
			Name: "diego_capacity",
		}}))
		g.Expect(docs[1].Labels).To(Equal(map[string]string{"deployment": "cf-abc-123", "product": "pas"}))
	})

	t.Run("documents expire after an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		time.Sleep(10 * time.Millisecond)
		g.Expect(d.All()).To(HaveLen(0))
	})
}
