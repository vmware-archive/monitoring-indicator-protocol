package indicator_status_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

func TestStatusMatcher(t *testing.T) {
	t.Run("returns undefined when there are no thresholds", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]v1.Threshold{}, []float64{1, 2, 3})
		g.Expect(status).To(Equal("UNDEFINED"))
	})

	t.Run("returns unknown when there are no values", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]v1.Threshold{{
			Level:    "warning",
			Operator: v1.LessThanOrEqualTo,
			Value:    0,
		}}, []float64{})
		g.Expect(status).To(Equal("UNKNOWN"))
	})

	t.Run("returns the threshold name when it's breached", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]v1.Threshold{{
			Level:    "warning",
			Operator: v1.LessThanOrEqualTo,
			Value:    0,
		}}, []float64{0})

		g.Expect(status).NotTo(BeNil())
		g.Expect(status).To(Equal("warning"))
	})

	t.Run("returns healthy if the only threshold provided isn't breached", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]v1.Threshold{{
			Level:    "warning",
			Operator: v1.LessThanOrEqualTo,
			Value:    0,
		}}, []float64{5})

		g.Expect(status).To(Equal("HEALTHY"))
	})

	t.Run("threshold operators", func(t *testing.T) {
		t.Run("greater than", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{{
				Level:    "breached",
				Operator: v1.GreaterThan,
				Value:    9,
			}, {
				Level:    "not_breached",
				Operator: v1.GreaterThan,
				Value:    10,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("greater than or equal", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{{
				Level:    "breached",
				Operator: v1.GreaterThanOrEqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: v1.GreaterThanOrEqualTo,
				Value:    9,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("less than", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{{
				Level:    "breached",
				Operator: v1.LessThan,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: v1.LessThan,
				Value:    9,
			}}, []float64{9})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("less than or equal", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{{
				Level:    "breached",
				Operator: v1.LessThanOrEqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: v1.LessThanOrEqualTo,
				Value:    9,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("equal to", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{{
				Level:    "breached",
				Operator: v1.EqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: v1.EqualTo,
				Value:    9,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("not equal to", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{{
				Level:    "breached",
				Operator: v1.NotEqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: v1.NotEqualTo,
				Value:    9,
			}}, []float64{9})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})
	})

	t.Run("threshold priority", func(t *testing.T) {
		t.Run("returns the first status in alphanumeric order if 'critical' or 'warning' haven't been breached", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{{
				Level:    "warning",
				Operator: v1.GreaterThan,
				Value:    9,
			}, {
				Level:    "critical",
				Operator: v1.GreaterThan,
				Value:    10,
			}, {
				Level:    "acceptable",
				Operator: v1.GreaterThan,
				Value:    8,
			}, {
				Level:    "1",
				Operator: v1.GreaterThan,
				Value:    1,
			},
			}, []float64{9})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("1"))
		})

		t.Run("returns the first status in alphanumeric order", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{
				{
					Level:    "abc",
					Operator: v1.GreaterThan,
					Value:    9,
				}, {
					Level:    "abcd",
					Operator: v1.GreaterThan,
					Value:    10,
				}, {
					Level:    "abc1",
					Operator: v1.GreaterThan,
					Value:    8,
				},
			}, []float64{20})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("abc"))
		})

		t.Run("returns critical if it has been breached", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{
				{
					Level:    "acceptable",
					Operator: v1.GreaterThan,
					Value:    8,
				}, {
					Level:    "warning",
					Operator: v1.GreaterThan,
					Value:    9,
				}, {
					Level:    "critical",
					Operator: v1.GreaterThan,
					Value:    10,
				},
			}, []float64{11})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("critical"))
		})

		t.Run("returns warning if it has been breached and critical has not", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]v1.Threshold{
				{
					Level:    "acceptable",
					Operator: v1.GreaterThan,
					Value:    8,
				}, {
					Level:    "warning",
					Operator: v1.GreaterThan,
					Value:    9,
				}, {
					Level:    "critical",
					Operator: v1.GreaterThan,
					Value:    10,
				},
			}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("warning"))
		})
	})

	t.Run("returns correct status when multiple values breach thresholds", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]v1.Threshold{
			{
				Level:    "acceptable",
				Operator: v1.GreaterThan,
				Value:    8,
			}, {
				Level:    "warning",
				Operator: v1.GreaterThan,
				Value:    9,
			}, {
				Level:    "critical",
				Operator: v1.GreaterThan,
				Value:    10,
			},
		}, []float64{9, 10, 12})

		g.Expect(status).NotTo(BeNil())
		g.Expect(status).To(Equal("critical"))
	})
}
