package test_fixtures

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func DefaultPresentation() v1.Presentation {
	return v1.Presentation{
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

func DefaultLayout(indicators []v1.IndicatorSpec) v1.Layout {
	indicatorNames := make([]string, 0, len(indicators))
	for _, i := range indicators {
		indicatorNames = append(indicatorNames, i.Name)
	}
	return v1.Layout{
		Title:       "",
		Description: "",
		Sections: []v1.Section{{
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

func DefaultAlert() v1.Alert {
	return v1.Alert{
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

func Indicator(name string, promql string) v1.Indicator {
	return v1.Indicator{
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Spec: v1.IndicatorSpec{
			Product: "CF",
			Name:    "test",
			PromQL:  promql,
			Thresholds: []v1.Threshold{{
				Level:    "critical",
				Operator: v1.LessThan,
				Value:    float64(0),
			}},
		},
	}
}
