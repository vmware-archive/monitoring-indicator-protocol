package indicator_status_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	"github.com/prometheus/common/model"
)

func TestQueryValues(t *testing.T) {
	t.Run("when Prometheus returns an empty vector, returns empty values", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])": model.Vector{},
		})

		values, err := indicator_status.QueryValues(fakeQueryClient, "rate(errors[5m])")

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(values).To(BeEmpty())
	})

	t.Run("when Prometheus returns a non-empty vector, returns non-empty values", func(t *testing.T) {
		g := NewGomegaWithT(t)

		expectedResponse := []float64{9, 10, 11, 12}

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])": vectorResponse(expectedResponse),
		})

		values, err := indicator_status.QueryValues(fakeQueryClient, "rate(errors[5m])")

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(values).To(Equal(expectedResponse))
	})

	t.Run("when Prometheus returns an error, returns the error", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(nil)
		fakeQueryClient.err = errors.New("test-error")

		values, err := indicator_status.QueryValues(fakeQueryClient, "rate(errors[5m])")

		g.Expect(err).To(HaveOccurred())
		g.Expect(values).To(BeNil())
	})

	t.Run("when Prometheus returns a non-vector value, returns the error", func(t *testing.T) {
		g := NewGomegaWithT(t)

		expectedResponse := &model.Scalar{}

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])": expectedResponse,
		})

		values, err := indicator_status.QueryValues(fakeQueryClient, "rate(errors[5m])")

		g.Expect(values).To(BeNil())
		g.Expect(err.Error()).To(Equal("could not assert result from prometheus as Vector: *model.Scalar"))
	})

	t.Run("it queries prometheus with proper context", func(t *testing.T) {
		g := NewGomegaWithT(t)

		expectedResponse := &model.Scalar{}

		fakeQueryClient := setupFakeQueryClientWithVectorResponses(map[string]model.Value{
			"rate(errors[5m])": expectedResponse,
		})

		indicator_status.QueryValues(fakeQueryClient, "rate(errors[5m])")

		ctx := fakeQueryClient.queryArgs[0].ctx
		g.Expect(ctx.Done()).To(BeClosed())
		_, hasDeadline := ctx.Deadline()
		g.Expect(hasDeadline).To(BeTrue())
	})
}
