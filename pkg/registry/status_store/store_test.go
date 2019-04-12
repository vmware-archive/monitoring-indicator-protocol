package status_store_test

import (
	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"

	"testing"
	"time"
)

func TestUpdatingStatus(t *testing.T) {

	now := time.Now()
	fakeClock := func() time.Time { return now }

	t.Run("it can find a single status that was updated", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := status_store.New(fakeClock)

		store.UpdateStatus(status_store.UpdateRequest{
			DocumentUID:   "abc-123",
			IndicatorName: "error_rate",
			Status:        test_fixtures.StrPtr("critical"),
		})

		store.UpdateStatus(status_store.UpdateRequest{
			DocumentUID:   "abc-123",
			IndicatorName: "latency",
			Status:        test_fixtures.StrPtr("critical"),
		})

		g.Expect(store.StatusFor("abc-123", "latency")).To(Equal(status_store.IndicatorStatus{
			DocumentUID:   "abc-123",
			IndicatorName: "latency",
			Status:        test_fixtures.StrPtr("critical"),
			UpdatedAt:     now,
		}))
	})

	t.Run("it returns an error if the status was never updated", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := status_store.New(fakeClock)

		_, err := store.StatusFor("abc-123", "latency")

		g.Expect(err).To(HaveOccurred())
	})

	t.Run("It can update an existing status", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := status_store.New(fakeClock)

		store.UpdateStatus(status_store.UpdateRequest{
			Status:        test_fixtures.StrPtr("healthy"),
			IndicatorName: "latency",
			DocumentUID:   "abc-123",
		})
		store.UpdateStatus(status_store.UpdateRequest{
			Status:        test_fixtures.StrPtr("critical"),
			IndicatorName: "latency",
			DocumentUID:   "abc-123",
		})

		g.Expect(store.StatusFor("abc-123", "latency")).To(Equal(status_store.IndicatorStatus{
			DocumentUID:   "abc-123",
			IndicatorName: "latency",
			Status:        test_fixtures.StrPtr("critical"),
			UpdatedAt:     now,
		}))
	})
}
