package indicator_test

import (
  . "github.com/onsi/gomega"
  "testing"

  "code.cloudfoundry.org/indicators/pkg/indicator"
)

func TestUpdateMetadata(t *testing.T) {
  t.Run("it replaces promql $EXPR with metadata tags", func(t *testing.T) {
    g := NewGomegaWithT(t)
    d, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product: well-performing-component
version: 0.0.1
metadata:
  deployment: well-performing-deployment

indicators:
- name: test_performance_indicator
  promql: query_metric{source_id="$deployment"}
`))
    g.Expect(err).ToNot(HaveOccurred())

    g.Expect(d).To(Equal(indicator.Document{
      APIVersion: "v0",
      Product:    "well-performing-component",
      Version:    "0.0.1",
      Metadata:   map[string]string{"deployment": "well-performing-deployment"},
      Indicators: []indicator.Indicator{
        {
          Name:   "test_performance_indicator",
          PromQL: `query_metric{source_id="well-performing-deployment"}`,
        },
      },
      Documentation: indicator.Documentation{},
    }))
  })
}
