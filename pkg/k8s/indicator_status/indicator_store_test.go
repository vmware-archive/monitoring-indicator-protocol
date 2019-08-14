package indicator_status_test

import (
	"bytes"
	"log"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestIndicatorStore(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	t.Run("Add", func(t *testing.T) {
		t.Run("adds an item to the store", func(t *testing.T) {
			g := NewGomegaWithT(t)
			indicator := test_fixtures.Indicator("an-indicator", "rate(love[5m])")
			store := indicator_status.NewIndicatorStore()
			store.Add(indicator)
			g.Expect(store.GetIndicators()).To(ConsistOf(indicator))
		})

		t.Run("it can handle concurrent adds to the store", func(t *testing.T) {
			g := NewGomegaWithT(t)
			indicator := test_fixtures.Indicator("an-indicator", "rate(love[5m])")
			indicator2 := test_fixtures.Indicator("another-indicator", "rate(love[8m])")
			store := indicator_status.NewIndicatorStore()
			go store.Add(indicator)
			go store.Add(indicator2)
			g.Eventually(store.GetIndicators).Should(ConsistOf(indicator, indicator2))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("removes an item from the store", func(t *testing.T) {
			g := NewGomegaWithT(t)
			indicator := test_fixtures.Indicator("an-indicator", "rate(love[5m])")
			store := indicator_status.NewIndicatorStore()
			store.Add(indicator)

			store.Delete(indicator)
			g.Expect(store.GetIndicators()).ToNot(ConsistOf(indicator))
		})

		t.Run("it can handle concurrent removes from the store", func(t *testing.T) {
			g := NewGomegaWithT(t)
			indicator := test_fixtures.Indicator("an-indicator", "rate(love[5m])")
			indicator2 := test_fixtures.Indicator("another-indicator", "rate(love[8m])")
			store := indicator_status.NewIndicatorStore()
			store.Add(indicator)
			store.Add(indicator2)
			go store.Delete(indicator)
			go store.Delete(indicator2)
			g.Eventually(store.GetIndicators).Should(HaveLen(0))
			g.Eventually(store.GetIndicators).ShouldNot(ConsistOf(indicator, indicator2))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("updates an item in the store", func(t *testing.T) {
			g := NewGomegaWithT(t)
			indicator := test_fixtures.Indicator("an-indicator", "rate(love[5m])")
			store := indicator_status.NewIndicatorStore()
			store.Add(indicator)

			updatedIndicator := test_fixtures.Indicator("an-indicator", "rate(more-love[5m]")
			store.Update(updatedIndicator)
			g.Expect(store.GetIndicators()[0].Spec.PromQL).To(Equal("rate(more-love[5m]"))
		})

		t.Run("it can handle concurrent updates", func(t *testing.T) {
			g := NewGomegaWithT(t)
			indicator := test_fixtures.Indicator("an-indicator", "rate(love[5m])")
			updated1 := test_fixtures.Indicator("an-indicator", "rate(love[8m])")
			updated2 := test_fixtures.Indicator("an-indicator", "rate(love[10m])")
			store := indicator_status.NewIndicatorStore()
			store.Add(indicator)

			go store.Update(updated1)
			go store.Update(updated2)

			g.Consistently(store.GetIndicators).Should(HaveLen(1))
		})
	})
}
