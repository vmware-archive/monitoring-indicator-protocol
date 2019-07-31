package status_store_test

import (
	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func TestFillingStatuses(t *testing.T) {
	fixedTime := time.Now()
	fakeClock := func() time.Time { return fixedTime }

	t.Run("fills in statuses based on what is in the store", func(t *testing.T) {
		g := NewGomegaWithT(t)

		store := status_store.New(fakeClock)

		store.UpdateStatus(status_store.UpdateRequest{
			DocumentUID:   "abc-be46e8db7a40d475c055db41ad1d55bbef19335f",
			IndicatorName: "error_rate",
			Status:        test_fixtures.StrPtr("critical"),
		})

		store.UpdateStatus(status_store.UpdateRequest{
			DocumentUID:   "abc-be46e8db7a40d475c055db41ad1d55bbef19335f",
			IndicatorName: "latency",
			Status:        test_fixtures.StrPtr("warning"),
		})

		store.UpdateStatus(status_store.UpdateRequest{
			DocumentUID:   "abc-foo6e8db7a40d475c055db41ad1d55bbef19335f",
			IndicatorName: "latency",
			Status:        test_fixtures.StrPtr("healthy"),
		})

		document := v1.IndicatorDocument{
			TypeMeta: metaV1.TypeMeta{
				Kind:       "IndicatorDocument",
				APIVersion: api_versions.V1,
			},
			ObjectMeta: metaV1.ObjectMeta{
				Labels: map[string]string{
					"source_id": "bar",
				},
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "abc",
					Version: "1.2.3",
				},
				Indicators: []v1.IndicatorSpec{{
					Name:   "error_rate",
					PromQL: "error_rate",
				}, {
					Name:   "latency",
					PromQL: "latency",
				}},
			},
		}

		store.FillStatuses(&document)

		g.Expect(document.Status).To(BeEquivalentTo(map[string]v1.IndicatorStatus{
			"error_rate": {
				Phase:     "critical",
				UpdatedAt: metaV1.Time{Time: fixedTime},
			},
			"latency": {
				Phase:     "warning",
				UpdatedAt: metaV1.Time{Time: fixedTime},
			},
		}))
	})
}
