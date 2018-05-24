package docs

import "code.cloudfoundry.org/cf-indicators/pkg/indicator"

type Documentation struct {
	Title string
	Description string
	Sections []Section
}

type Section struct {
	Title string
	Description string
	Metrics []indicator.Metric
	Indicators []indicator.Indicator
}
