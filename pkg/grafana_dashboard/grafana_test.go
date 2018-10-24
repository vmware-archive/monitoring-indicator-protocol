package grafana_dashboard_test

import (
	"code.cloudfoundry.org/indicators/pkg/grafana_dashboard"
	. "github.com/onsi/gomega"
	"testing"

	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func TestDocumentToDashboard(t *testing.T) {
	g := NewGomegaWithT(t)

	document := indicator.Document{
		Indicators:    []indicator.Indicator{{
			Name:          "test_indicator",
			PromQL:        `sum_over_time(gorouter_latency_ms[30m])`,
			Thresholds:    []indicator.Threshold{{
				Level:    "critical",
				Operator: indicator.GreaterThan,
				Value:    1000,
			}},
			SLO:           0.999,
		}},
		Documentation: indicator.Documentation{
			Title:       "Indicator Test Dashboard",
		},
	}

	text, err := grafana_dashboard.DocumentToDashboard(document)
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it outputs a grafana dashboard definition", func(t *testing.T) {
		g := NewGomegaWithT(t)
		g.Expect(text).To(ContainSubstring(`"title":"test_indicator"`))
		g.Expect(text).To(ContainSubstring(`"expr":"sum_over_time(gorouter_latency_ms[30m])"`))
	})
}
