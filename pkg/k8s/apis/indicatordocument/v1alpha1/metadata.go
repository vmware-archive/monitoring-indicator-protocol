package v1alpha1

import (
	"fmt"
	"regexp"
)

func (id *IndicatorDocument) OverrideMetadata(overrides map[string]string) {
	for k, v := range overrides {
		id.ObjectMeta.Labels[k] = v
	}
}

func (id *IndicatorDocument) Interpolate() {
	for k, v := range id.ObjectMeta.Labels {
		regString := fmt.Sprintf(`(?i)\$%s\b`, k)
		reg := regexp.MustCompile(regString)

		for i, indicator := range id.Spec.Indicators {
			id.Spec.Indicators[i].PromQL = reg.ReplaceAllString(indicator.PromQL, v)
		}
	}
}
