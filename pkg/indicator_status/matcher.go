package indicator_status

import (
	"sort"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func Match(thresholds []registry.APIV0Threshold, values []float64) *string {
	if len(thresholds) == 0 {
		return nil
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
		return nil
	}

	status := selectThreshold(breachedThresholdLevels)
	return &status
}

func selectThreshold(thresholdLevels []string) string {
	sort.Sort(byThresholdLevel(thresholdLevels))
	return thresholdLevels[0]
}

func isBreached(threshold registry.APIV0Threshold, value float64) bool {
	switch threshold.Operator {
	case "lte":
		return value <= threshold.Value
	case "lt":
		return value < threshold.Value
	case "gte":
		return value >= threshold.Value
	case "gt":
		return value > threshold.Value
	case "neq":
		return value != threshold.Value
	case "eq":
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
