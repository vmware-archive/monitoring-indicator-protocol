package evaluator_test

import (
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"github.com/cloudfoundry-incubator/event-producer/pkg/evaluator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evaluator", func() {
	It("returns a list of events that satisfy the thresholds", func() {
		events := evaluator.GetSatisfiedEvents(&logcache_v1.PromQL_QueryResult{
			Result: &logcache_v1.PromQL_QueryResult_Vector{
				Vector: &logcache_v1.PromQL_Vector{
					Samples: []*logcache_v1.PromQL_Sample{
						{
							Metric: map[string]string{"ip": "10.0.0.1"},
							Point: &logcache_v1.PromQL_Point{
								Time:  12345000,
								Value: 999,
							},
						},
						{
							Metric: map[string]string{"ip": "10.0.0.2"},
							Point: &logcache_v1.PromQL_Point{
								Time:  12345000,
								Value: 1001,
							},
						},
					},
				},
			},
		}, []kpi.Threshold{{
			Level:    "Critical",
			Operator: kpi.LessThan,
			Value:    1000,
		}})

		Expect(events).To(ConsistOf(kpi.Event{
			Tags:           map[string]string{"ip": "10.0.0.1"},
			Value:          999,
			ThresholdLevel: "Critical",
			ThresholdValue: 1000,
		}))
	})

	Context("Operators", func() {
		It("handles LessThan", func() {
			events := getEvents(kpi.LessThan, 1000, []float64{999, 1000})
			Expect(values(events)).To(ContainElement(float64(999)))
		})

		It("handles LessThanOrEqualTo", func() {
			events := getEvents(kpi.LessThanOrEqualTo, 1000, []float64{1000, 1000.1})
			Expect(values(events)).To(ContainElement(float64(1000)))
		})

		It("handles EqualTo", func() {
			events := getEvents(kpi.EqualTo, 1000, []float64{999, 1000, 1000.1})
			Expect(values(events)).To(ContainElement(float64(1000)))
		})

		It("handles NotEqualTo", func() {
			events := getEvents(kpi.NotEqualTo, 1000, []float64{999, 1000})
			Expect(values(events)).To(ContainElement(float64(999)))
		})

		It("handles GreaterThanOrEqualTo", func() {
			events := getEvents(kpi.GreaterThanOrEqualTo, 1000, []float64{999, 1000})
			Expect(values(events)).To(ContainElement(float64(1000)))
		})

		It("handles GreaterThan", func() {
			events := getEvents(kpi.GreaterThan, 1000, []float64{1000, 1000.1})
			Expect(values(events)).To(ContainElement(float64(1000.1)))
		})
	})
})

func values(events []kpi.Event) []float64 {
	var vals []float64
	for _, event := range events {
		vals = append(vals, event.Value)
	}

	return vals
}

func getEvents(operatorType kpi.OperatorType, thresholdValue float64, values []float64) []kpi.Event {
	samples := make([]*logcache_v1.PromQL_Sample, 0)
	for _, value := range values {
		samples = append(samples, &logcache_v1.PromQL_Sample{
			Metric: map[string]string{"ip": "10.0.0.1"},
			Point: &logcache_v1.PromQL_Point{
				Time:  1525797112000000000,
				Value: value,
			},
		})
	}

	return evaluator.GetSatisfiedEvents(&logcache_v1.PromQL_QueryResult{
		Result: &logcache_v1.PromQL_QueryResult_Vector{
			Vector: &logcache_v1.PromQL_Vector{
				Samples: samples,
			},
		},
	}, []kpi.Threshold{{
		Level:    "Critical",
		Operator: operatorType,
		Value:    thresholdValue,
	}})
}
