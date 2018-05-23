package evaluator_test

import (
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"github.com/cloudfoundry-incubator/event-producer/pkg/evaluator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/indicator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/producer"

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
		}, []indicator.Threshold{
			{
				Level:    "Critical",
				Operator: indicator.EqualTo,
				Value:    1001,
			}, {
				Level:    "Warning",
				Operator: indicator.LessThan,
				Value:    1000,
			},
		})

		Expect(events).To(ConsistOf(producer.Event{
			Tags:           map[string]string{"event_ip": "10.0.0.1"},
			Value:          999,
			ThresholdLevel: "Warning",
			ThresholdValue: 1000,
		}, producer.Event{
			Tags:           map[string]string{"event_ip": "10.0.0.2"},
			Value:          1001,
			ThresholdLevel: "Critical",
			ThresholdValue: 1001,
		}))
	})

	It("prefixes tags with 'event_'", func() {
		events := evaluator.GetSatisfiedEvents(&logcache_v1.PromQL_QueryResult{
			Result: &logcache_v1.PromQL_QueryResult_Vector{
				Vector: &logcache_v1.PromQL_Vector{
					Samples: []*logcache_v1.PromQL_Sample{
						{
							Metric: map[string]string{"tag": "test1", "tag2": "test_tag"},
							Point: &logcache_v1.PromQL_Point{
								Time:  12345000,
								Value: 1000,
							},
						},
					},
				},
			},
		}, []indicator.Threshold{
			{
				Level:    "Critical",
				Operator: indicator.EqualTo,
				Value:    1000,
			},
		})

		Expect(events).To(ConsistOf(producer.Event{
			Tags:           map[string]string{"event_tag": "test1", "event_tag2": "test_tag"},
			Value:          1000,
			ThresholdLevel: "Critical",
			ThresholdValue: 1000,
		}))
	})

	Context("Operators", func() {
		It("handles LessThan", func() {
			events := getEvents(indicator.LessThan, 1000, []float64{999, 1000})
			Expect(values(events)).To(ContainElement(float64(999)))
		})

		It("handles LessThanOrEqualTo", func() {
			events := getEvents(indicator.LessThanOrEqualTo, 1000, []float64{1000, 1000.1})
			Expect(values(events)).To(ContainElement(float64(1000)))
		})

		It("handles EqualTo", func() {
			events := getEvents(indicator.EqualTo, 1000, []float64{999, 1000, 1000.1})
			Expect(values(events)).To(ContainElement(float64(1000)))
		})

		It("handles NotEqualTo", func() {
			events := getEvents(indicator.NotEqualTo, 1000, []float64{999, 1000})
			Expect(values(events)).To(ContainElement(float64(999)))
		})

		It("handles GreaterThanOrEqualTo", func() {
			events := getEvents(indicator.GreaterThanOrEqualTo, 1000, []float64{999, 1000})
			Expect(values(events)).To(ContainElement(float64(1000)))
		})

		It("handles GreaterThan", func() {
			events := getEvents(indicator.GreaterThan, 1000, []float64{1000, 1000.1})
			Expect(values(events)).To(ContainElement(float64(1000.1)))
		})
	})
})

func values(events []producer.Event) []float64 {
	var vals []float64
	for _, event := range events {
		vals = append(vals, event.Value)
	}

	return vals
}

func getEvents(operatorType indicator.OperatorType, thresholdValue float64, values []float64) []producer.Event {
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
	}, []indicator.Threshold{{
		Level:    "Critical",
		Operator: operatorType,
		Value:    thresholdValue,
	}})
}
