package test_fixtures

import (
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func DefaultPresentation() *indicator.Presentation {
	return &indicator.Presentation{
		ChartType:    "step",
		CurrentValue: false,
		Frequency:    0,
		Labels:       nil,
	}
}

func DefaultAPIV0Presentation() *registry.APIV0Presentation {
	return &registry.APIV0Presentation{
		ChartType:    "step",
		CurrentValue: false,
		Frequency:    0,
		Labels:       nil,
	}
}

func DefaultLayout(indicators []indicator.Indicator) indicator.Layout {
	return indicator.Layout{
		Title:       "",
		Description: "",
		Sections: []indicator.Section{{
			Title:       "",
			Description: "",
			Indicators:  indicators,
		}},
		Owner: "",
	}
}

func DefaultAPIV0Layout(indicatorNames []string) registry.APIV0Layout {
	return registry.APIV0Layout{
		Title:       "",
		Description: "",
		Sections: []registry.APIV0Section{{
			Title:       "",
			Description: "",
			Indicators:  indicatorNames,
		}},
		Owner: "",
	}
}

func DefaultAlert() indicator.Alert {
	return indicator.Alert{
		For:  "1m",
		Step: "1m",
	}
}

func DefaultAPIV0Alert() registry.APIV0Alert {
	return registry.APIV0Alert{
		For:  "1m",
		Step: "1m",
	}
}
