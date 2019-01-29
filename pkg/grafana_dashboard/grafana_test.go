package grafana_dashboard_test

import (
  "bytes"
  "log"
  "testing"

  . "github.com/onsi/gomega"

  "github.com/pivotal/indicator-protocol/pkg/grafana_dashboard"
  "github.com/pivotal/indicator-protocol/pkg/indicator"
)

func TestDocumentToDashboard(t *testing.T) {
  buffer := bytes.NewBuffer(nil)
  log.SetOutput(buffer)

  g := NewGomegaWithT(t)

  document := indicator.Document{
    Indicators: []indicator.Indicator{
      {
        Name:   "test_indicator",
        PromQL: `sum_over_time(gorouter_latency_ms[30m])`,
        Thresholds: []indicator.Threshold{{
          Level:    "critical",
          Operator: indicator.GreaterThan,
          Value:    1000,
        }, {
          Level:    "warning",
          Operator: indicator.NotEqualTo,
          Value:    17,
        }, {
          Level:    "warning",
          Operator: indicator.GreaterThanOrEqualTo,
          Value:    700,
        }, {
          Level:    "uh-oh!",
          Operator: indicator.LessThan,
          Value:    500,
        }},
      },
      {
        Name:       "second_test_indicator",
        PromQL:     `rate(gorouter_requests[1m])`,
        Thresholds: []indicator.Threshold{},
      },
    },
    Layout: indicator.Layout{
      Title: "Indicator Test Dashboard",
    },
  }

  text, err := grafana_dashboard.DocumentToDashboard(document)
  g.Expect(err).ToNot(HaveOccurred())

  t.Run("it outputs a grafana dashboard definition", func(t *testing.T) {
    g := NewGomegaWithT(t)
    g.Expect(text).To(ContainSubstring(`"title": "Indicator Test Dashboard"`))
    g.Expect(text).To(ContainSubstring(`"id": null`))
    g.Expect(text).To(ContainSubstring(`"expr": "sum_over_time(gorouter_latency_ms[30m])"`))
  })

  t.Run("it has thresholds on graphs", func(t *testing.T) {
    g := NewGomegaWithT(t)
    g.Expect(text).To(ContainSubstring(`"thresholds": [`))
    g.Expect(text).To(ContainSubstring(`"value":1000, "colorMode":"critical", "op":"gt"`))
    g.Expect(text).To(ContainSubstring(`"value":700, "colorMode":"warning", "op":"gt"`))
    g.Expect(text).ToNot(ContainSubstring(`"colorMode":""`))
  })

  t.Run("it skips thresholds with unrepresentable operators", func(t *testing.T) {
    g := NewGomegaWithT(t)
    g.Expect(text).ToNot(ContainSubstring(`"value":17, "colorMode":"warning"`))
  })

  t.Run("it uses custom coloring in thresholds with unknown level", func(t *testing.T) {
    g := NewGomegaWithT(t)
    g.Expect(text).To(ContainSubstring(`"value":500, "colorMode":"custom", "op":"lt"`))
  })
}
