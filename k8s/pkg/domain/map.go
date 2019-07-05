package domain

import (
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func Map(i *v1alpha1.IndicatorDocument) indicator.Document {
	indicators := mapToDomainIndicators(i.Spec.Indicators)
	return indicator.Document{
		Product: indicator.Product{
			Name:    i.Spec.Product.Name,
			Version: i.Spec.Product.Version,
		},
		Metadata:   i.Labels,
		APIVersion: i.APIVersion,
		Indicators: indicators,

		Layout: indicator.Layout{
			Title:       i.Spec.Layout.Title,
			Description: i.Spec.Layout.Description,
			Sections:    mapToDomainSections(i.Spec.Layout.Sections, indicators),
			Owner:       i.Spec.Layout.Owner,
		},
	}
}

func mapToDomainSections(sections []v1alpha1.Section, indicators []indicator.Indicator) []indicator.Section {
	domainSections := make([]indicator.Section, 0, len(sections))

	for _, s := range sections {
		domainSections = append(domainSections, indicator.Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  s.Indicators,
		})
	}

	return domainSections
}

func findIndicators(names []string, indicators []indicator.Indicator) []indicator.Indicator {
	matchedIndicators := make([]indicator.Indicator, 0)

	for _, n := range names {
		for _, i := range indicators {
			if i.Name == n {
				matchedIndicators = append(matchedIndicators, i)
			}
		}
	}

	return matchedIndicators
}

func mapToDomainIndicators(ids []v1alpha1.IndicatorSpec) []indicator.Indicator {
	indicators := make([]indicator.Indicator, 0, len(ids))
	for _, i := range ids {
		indicators = append(indicators, ToDomainIndicator(i))
	}
	return indicators
}

func ToDomainIndicator(i v1alpha1.IndicatorSpec) indicator.Indicator {
	return indicator.Indicator{
		Name:          i.Name,
		PromQL:        i.Promql,
		Alert:         toDomainAlert(i.Alert),
		Thresholds:    MapToDomainThreshold(i.Thresholds),
		Documentation: i.Documentation,
		Presentation:  toDomainPresentation(i.Presentation),
	}
}

func toDomainPresentation(presentation v1alpha1.Presentation) indicator.Presentation {
	return indicator.Presentation{
		ChartType:    presentation.ChartType,
		CurrentValue: presentation.CurrentValue,
		Frequency:    presentation.Frequency,
		Labels:       presentation.Labels,
	}
}

func toDomainAlert(a v1alpha1.Alert) indicator.Alert {
	return indicator.Alert{
		For:  a.For,
		Step: a.Step,
	}
}

func MapToDomainThreshold(ths []v1alpha1.Threshold) []indicator.Threshold {
	thresholds := make([]indicator.Threshold, 0, len(ths))
	for _, t := range ths {
		op := resolveOperator(t)
		thresholds = append(thresholds, indicator.Threshold{
			Level:    t.Level,
			Operator: op,
			Value:    t.Value,
		})
	}
	return thresholds
}

func resolveOperator(t v1alpha1.Threshold) indicator.OperatorType {
	switch {
	case t.Operator == "lt":
		return indicator.LessThan
	case t.Operator == "lte":
		return indicator.LessThanOrEqualTo
	case t.Operator == "eq":
		return indicator.EqualTo
	case t.Operator == "neq":
		return indicator.NotEqualTo
	case t.Operator == "gte":
		return indicator.GreaterThanOrEqualTo
	case t.Operator == "gt":
		return indicator.GreaterThan
	}

	return indicator.Undefined
}
