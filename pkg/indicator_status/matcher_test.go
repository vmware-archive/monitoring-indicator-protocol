package indicator_status_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
)

func TestStatusMatcher(t *testing.T) {
	t.Run("returns undefined when there are no thresholds", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]indicator.Threshold{}, []float64{1, 2, 3})
		g.Expect(status).To(Equal("UNDEFINED"))
	})

	t.Run("returns healthy when there are no values", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]indicator.Threshold{{
			Level:    "warning",
			Operator: indicator.LessThanOrEqualTo,
			Value:    0,
		}}, []float64{})
		g.Expect(status).To(Equal("HEALTHY"))
	})

	t.Run("returns the threshold name when it's breached", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]indicator.Threshold{{
			Level:    "warning",
			Operator: indicator.LessThanOrEqualTo,
			Value:    0,
		}}, []float64{0})

		g.Expect(status).NotTo(BeNil())
		g.Expect(status).To(Equal("warning"))
	})

	t.Run("returns healthy if the only threshold provided isn't breached", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]indicator.Threshold{{
			Level:    "warning",
			Operator: indicator.LessThanOrEqualTo,
			Value:    0,
		}}, []float64{5})

		g.Expect(status).To(Equal("HEALTHY"))
	})

	t.Run("threshold operators", func(t *testing.T) {
		t.Run("greater than", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{{
				Level:    "breached",
				Operator: indicator.GreaterThan,
				Value:    9,
			}, {
				Level:    "not_breached",
				Operator: indicator.GreaterThan,
				Value:    10,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("greater than or equal", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{{
				Level:    "breached",
				Operator: indicator.GreaterThanOrEqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: indicator.GreaterThanOrEqualTo,
				Value:    9,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("less than", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{{
				Level:    "breached",
				Operator: indicator.LessThan,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: indicator.LessThan,
				Value:    9,
			}}, []float64{9})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("less than or equal", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{{
				Level:    "breached",
				Operator: indicator.LessThanOrEqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: indicator.LessThanOrEqualTo,
				Value:    9,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("equal to", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{{
				Level:    "breached",
				Operator: indicator.EqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: indicator.EqualTo,
				Value:    9,
			}}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})

		t.Run("not equal to", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{{
				Level:    "breached",
				Operator: indicator.NotEqualTo,
				Value:    10,
			}, {
				Level:    "not_breached",
				Operator: indicator.NotEqualTo,
				Value:    9,
			}}, []float64{9})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("breached"))
		})
	})

	t.Run("threshold priority", func(t *testing.T) {
		t.Run("returns the first status in alphanumeric order if 'critical' or 'warning' haven't been breached", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{{
				Level:    "warning",
				Operator: indicator.GreaterThan,
				Value:    9,
			}, {
				Level:    "critical",
				Operator: indicator.GreaterThan,
				Value:    10,
			}, {
				Level:    "acceptable",
				Operator: indicator.GreaterThan,
				Value:    8,
			}, {
				Level:    "1",
				Operator: indicator.GreaterThan,
				Value:    1,
			},
			}, []float64{9})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("1"))
		})

		t.Run("returns the first status in alphanumeric order", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{
				{
					Level:    "abc",
					Operator: indicator.GreaterThan,
					Value:    9,
				}, {
					Level:    "abcd",
					Operator: indicator.GreaterThan,
					Value:    10,
				}, {
					Level:    "abc1",
					Operator: indicator.GreaterThan,
					Value:    8,
				},
			}, []float64{20})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("abc"))
		})

		t.Run("returns critical if it has been breached", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{
				{
					Level:    "acceptable",
					Operator: indicator.GreaterThan,
					Value:    8,
				}, {
					Level:    "warning",
					Operator: indicator.GreaterThan,
					Value:    9,
				}, {
					Level:    "critical",
					Operator: indicator.GreaterThan,
					Value:    10,
				},
			}, []float64{11})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("critical"))
		})

		t.Run("returns warning if it has been breached and critical has not", func(t *testing.T) {
			g := NewGomegaWithT(t)

			status := indicator_status.Match([]indicator.Threshold{
				{
					Level:    "acceptable",
					Operator: indicator.GreaterThan,
					Value:    8,
				}, {
					Level:    "warning",
					Operator: indicator.GreaterThan,
					Value:    9,
				}, {
					Level:    "critical",
					Operator: indicator.GreaterThan,
					Value:    10,
				},
			}, []float64{10})

			g.Expect(status).NotTo(BeNil())
			g.Expect(status).To(Equal("warning"))
		})
	})

	t.Run("returns correct status when multiple values breach thresholds", func(t *testing.T) {
		g := NewGomegaWithT(t)

		status := indicator_status.Match([]indicator.Threshold{
			{
				Level:    "acceptable",
				Operator: indicator.GreaterThan,
				Value:    8,
			}, {
				Level:    "warning",
				Operator: indicator.GreaterThan,
				Value:    9,
			}, {
				Level:    "critical",
				Operator: indicator.GreaterThan,
				Value:    10,
			},
		}, []float64{9, 10, 12})

		g.Expect(status).NotTo(BeNil())
		g.Expect(status).To(Equal("critical"))
	})
}
