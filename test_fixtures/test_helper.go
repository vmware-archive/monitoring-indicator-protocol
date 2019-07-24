package test_fixtures

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func DefaultPresentation() v1alpha1.Presentation {
	return v1alpha1.Presentation{
		ChartType:    "step",
		CurrentValue: false,
		Frequency:    0,
		Labels:       nil,
		Units:        "",
	}
}

func DefaultAPIPresentationResponse() registry.APIPresentationResponse {
	return registry.APIPresentationResponse{
		ChartType:    "step",
		CurrentValue: false,
		Frequency:    0,
		Labels:       nil,
		Units:        "",
	}
}

func DefaultLayout(indicators []v1alpha1.IndicatorSpec) v1alpha1.Layout {
	indicatorNames := make([]string, 0, len(indicators))
	for _, i := range indicators {
		indicatorNames = append(indicatorNames, i.Name)
	}
	return v1alpha1.Layout{
		Title:       "",
		Description: "",
		Sections: []v1alpha1.Section{{
			Title:       "",
			Description: "",
			Indicators:  indicatorNames,
		}},
		Owner: "",
	}
}

func DefaultAPILayoutResponse(indicatorNames []string) registry.APILayoutResponse {
	return registry.APILayoutResponse{
		Title:       "",
		Description: "",
		Sections: []registry.APISectionResponse{{
			Title:       "",
			Description: "",
			Indicators:  indicatorNames,
		}},
		Owner: "",
	}
}

func DefaultAlert() v1alpha1.Alert {
	return v1alpha1.Alert{
		For:  "1m",
		Step: "1m",
	}
}

func DefaultAPIAlertResponse() registry.APIAlertResponse {
	return registry.APIAlertResponse{
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
			PromQL:  promql,
			Thresholds: []v1alpha1.Threshold{{
				Level:    "critical",
				Operator: v1alpha1.LessThan,
				Value:    float64(0),
			}},
		},
	}
}
