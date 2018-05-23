package docs

import "github.com/cloudfoundry-incubator/event-producer/pkg/indicator"

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
