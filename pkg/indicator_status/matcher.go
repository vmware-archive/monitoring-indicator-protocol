package indicator_status

import (
	"sort"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

const healthy = "HEALTHY"
const undefined = "UNDEFINED"
const unknown = "UNKNOWN"

// Match takes thresholds and values and determines what threshold has been
// breached. It returns nil if nothing was breached.
func Match(thresholds []indicator.Threshold, values []float64) string {
	if len(thresholds) == 0 {
		return undefined
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

func isBreached(threshold indicator.Threshold, value float64) bool {
	switch threshold.Operator {
	case indicator.LessThanOrEqualTo:
		return value <= threshold.Value
	case indicator.LessThan:
		return value < threshold.Value
	case indicator.GreaterThanOrEqualTo:
		return value >= threshold.Value
	case indicator.GreaterThan:
		return value > threshold.Value
	case indicator.NotEqualTo:
		return value != threshold.Value
	case indicator.EqualTo:
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
