package docs

import "code.cloudfoundry.org/cf-indicators/pkg/indicator"

type Documentation struct {
	Title       string
	Description string
	Sections    []Section
	Owner       string
}

type Section struct {
	Title       string
	Description string
	Metrics     []indicator.Metric
	Indicators  []indicator.Indicator
}
