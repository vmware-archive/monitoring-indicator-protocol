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
		Metadata: i.Labels,
		Indicators: indicators,

		// TODO: add layouts correctly
		Layout: indicator.Layout{
			Title:       i.Spec.Layout.Title,
		},
	}
}

func mapToDomainIndicators(ids []v1alpha1.Indicator) []indicator.Indicator {
	indicators := make([]indicator.Indicator, 0, len(ids))
	for _, i :=range ids {
		indicators = append(indicators, indicator.Indicator{
			Name:          i.Name,
			PromQL:        i.Promql,
			Thresholds:    mapToDomainThreshold(i.Thresholds),
			Documentation: i.Documentation,
		})
	}
	return indicators
}

func mapToDomainThreshold(ths []v1alpha1.Threshold) []indicator.Threshold {
	thresholds := make([]indicator.Threshold, 0, len(ths))
	for _, t := range ths {
		op, val := resolveOperator(t)
		thresholds = append(thresholds, indicator.Threshold{
			Level:    t.Level,
			Operator: op,
			Value:    val,
		})
	}
	return thresholds
}

func resolveOperator(t v1alpha1.Threshold) (indicator.OperatorType, float64) {
	switch {
	case t.Lt != nil:
		return indicator.LessThan, *t.Lt
	case t.Lte != nil:
		return indicator.LessThanOrEqualTo, *t.Lte
	case t.Eq != nil:
		return indicator.EqualTo, *t.Eq
	case t.Neq != nil:
		return indicator.NotEqualTo, *t.Neq
	case t.Gte != nil:
		return indicator.GreaterThanOrEqualTo, *t.Gte
	case t.Gt != nil:
		return indicator.GreaterThan, *t.Gt
	}

	return indicator.LessThan, 0
}
