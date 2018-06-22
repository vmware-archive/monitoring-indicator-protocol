package registry_test

import (
	"testing"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cf-indicators/pkg/registry"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
)

func TestInsertDocument(t *testing.T) {
	t.Run("it saves documents sent to it", func(t *testing.T) {
		g := NewGomegaWithT(t)

		d := registry.NewDocumentStore()
		d.Insert(
			map[string]string{"test_label": "test_value"},
			[]indicator.Indicator{{
				Name:  "test_name",
				Title: "test_title",
			}},
		)

		g.Expect(d.All()).To(ContainElement(registry.Document{
			Indicators: []indicator.Indicator{{
				Name:  "test_name",
				Title: "test_title",
			}},
			Labels: map[string]string{"test_label": "test_value"},
		}))
	})

}
