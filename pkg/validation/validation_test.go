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

func TestFormatQuery(t *testing.T) {

	var characterConversions = []struct {
		input       indicator.Metric
		expectation string
	}{
		{
			input:       indicator.Metric{SourceID: "router", Name: "uaa.latency"},
			expectation: `uaa_latency{source_id="router",deployment="cf"}[1m]`,
		},
		{
			input:       indicator.Metric{SourceID: "router", Name: `uaa/latency\a`},
			expectation: `uaa_latency_a{source_id="router",deployment="cf"}[1m]`,
		},
		{
			input:       indicator.Metric{SourceID: "router", Name: "uaa-latency"},
			expectation: `uaa_latency{source_id="router",deployment="cf"}[1m]`,
		},
	}

	for _, cc := range characterConversions {
		t.Run(cc.input.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(validation.FormatQuery(cc.input, "cf")).To(Equal(cc.expectation))
		})
	}
}
