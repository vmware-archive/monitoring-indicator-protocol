package validation_test

import (
	"testing"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cf-indicators/pkg/validation"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
)


// TODO: extracted from main, needs unit tests
func TestVerifyMetric(t *testing.T) {

}

// TODO: extracted from main, needs unit tests
func TestFormatQuery(t *testing.T) {

	t.Run("dots are converted to underscores", func(t *testing.T) {
		g := NewGomegaWithT(t)

		metric := indicator.Metric{
			SourceID:    "router",
			Name:        "uaa.latency",
		}

		g.Expect(validation.FormatQuery(metric, "cf")).To(Equal(`uaa_latency{source_id="router",deployment="cf"}[1m]`))
	})
}
