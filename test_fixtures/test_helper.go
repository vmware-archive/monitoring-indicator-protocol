package test_fixtures

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func DefaultPresentation() indicator.Presentation {
	return indicator.Presentation{
		ChartType:    "step",
		CurrentValue: false,
		Frequency:    0,
		Labels:       nil,
		Units:        "",
	}
}

func DefaultAPIV0Presentation() registry.APIV0Presentation {
	return registry.APIV0Presentation{
		ChartType:    "step",
		CurrentValue: false,
		Frequency:    0,
		Labels:       nil,
		Units:        "",
	}
}

func DefaultLayout(indicators []indicator.Indicator) indicator.Layout {
	indicatorNames := make([]string, 0, len(indicators))
	for _, i := range indicators {
		indicatorNames = append(indicatorNames, i.Name)
	}
	return indicator.Layout{
		Title:       "",
		Description: "",
		Sections: []indicator.Section{{
			Title:       "",
			Description: "",
			Indicators:  indicatorNames,
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

func StrPtr(s string) *string {
	return &s
}

func Indicator(name string, promql string) v1alpha1.Indicator {
	return v1alpha1.Indicator{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.IndicatorSpec{
			Product: "CF",
			Name:    "test",
			Promql:  promql,
			Thresholds: []v1alpha1.Threshold{{
				Level:    "critical",
				Operator: "lt",
				Value:    float64(0),
			}},
		},
	}
}
