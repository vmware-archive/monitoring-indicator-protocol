package evaluator

import (
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"
)

func GetSatisfiedEvents(result *logcache_v1.PromQL_QueryResult, thresholds []kpi.Threshold) []kpi.Event {
	events := make([]kpi.Event, 0)

	for _, threshold := range thresholds {
		for _, tuple := range convertToResultTuple(result) {
			if thresholdSatisfied(threshold, tuple.point) {
				events = append(events, kpi.Event{
					Tags:           tuple.tags,
					Value:          tuple.point.Value,
					ThresholdLevel: threshold.Level,
					ThresholdValue: threshold.Value,
				})
			}
		}
	}

	return events
}

func thresholdSatisfied(threshold kpi.Threshold, point *logcache_v1.PromQL_Point) bool {
	switch threshold.Operator {
	case kpi.LessThan:
		if point.Value < threshold.Value {
			return true
		}
	case kpi.LessThanOrEqualTo:
		if point.Value <= threshold.Value {
			return true
		}
	case kpi.EqualTo:
		if point.Value == threshold.Value {
			return true
		}
	case kpi.NotEqualTo:
		if point.Value != threshold.Value {
			return true
		}
	case kpi.GreaterThanOrEqualTo:
		if point.Value >= threshold.Value {
			return true
		}
	case kpi.GreaterThan:
		if point.Value > threshold.Value {
			return true
		}
	}

	return false
}

type resultTuple struct {
	tags  map[string]string
	point *logcache_v1.PromQL_Point
}

func convertToResultTuple(result *logcache_v1.PromQL_QueryResult) []resultTuple {
	var out []resultTuple
	switch r := result.GetResult().(type) {
	case *logcache_v1.PromQL_QueryResult_Vector:
		for _, sample := range r.Vector.GetSamples() {
			out = append(out, resultTuple{
				tags:  convertMetricTagNames(sample.GetMetric()),
				point: sample.GetPoint(),
			})
		}
	}

	return out
}

func convertMetricTagNames(tags map[string]string) map[string]string {
	convertedTags := make(map[string]string, len(tags))
	for tag, val := range tags {
		convertedTags["event_"+tag] = val
	}

	return convertedTags
}
