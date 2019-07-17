package indicator_status

import (
	"sort"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

const healthy = "HEALTHY"
const Undefined = "UNDEFINED"
const unknown = "UNKNOWN"

// Match takes thresholds and values and determines what threshold has been
// breached. It returns nil if nothing was breached.
func Match(thresholds []v1alpha1.Threshold, values []float64) string {
	if len(thresholds) == 0 {
		return Undefined
	}
	if len(values) == 0 {
		return unknown
	}
	var breachedThresholdLevels []string
	for _, threshold := range thresholds {
		for _, value := range values {
			if isBreached(threshold, value) {
				breachedThresholdLevels = append(breachedThresholdLevels, threshold.Level)
			}
		}
	}

	if len(breachedThresholdLevels) == 0 {
		return healthy
	}

	return selectThreshold(breachedThresholdLevels)
}

func selectThreshold(thresholdLevels []string) string {
	sort.Sort(byThresholdLevel(thresholdLevels))
	return thresholdLevels[0]
}

func isBreached(threshold v1alpha1.Threshold, value float64) bool {
	switch threshold.Operator {
	case v1alpha1.LessThanOrEqualTo:
		return value <= threshold.Value
	case v1alpha1.LessThan:
		return value < threshold.Value
	case v1alpha1.GreaterThanOrEqualTo:
		return value >= threshold.Value
	case v1alpha1.GreaterThan:
		return value > threshold.Value
	case v1alpha1.NotEqualTo:
		return value != threshold.Value
	case v1alpha1.EqualTo:
		return value == threshold.Value

	default:
		return false
	}
}

type byThresholdLevel []string

func (s byThresholdLevel) Len() int {
	return len(s)
}
func (s byThresholdLevel) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byThresholdLevel) Less(i, j int) bool {
	if s[i] == "critical" {
		return true
	}
	if s[i] == "warning" && s[j] != "critical" {
		return true
	}
	return sort.StringSlice(s).Less(i, j)
}
