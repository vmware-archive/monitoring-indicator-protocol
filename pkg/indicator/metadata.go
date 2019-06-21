package indicator

import (
	"fmt"
	"regexp"
)

func (document *Document) OverrideMetadata(overrides map[string]string) {
	for k, v := range overrides {
		document.Metadata[k] = v
	}
}

func (document *Document) Interpolate() {
	for k, v := range document.Metadata {
		regString := fmt.Sprintf(`(?i)\$%s\b`, k)
		reg := regexp.MustCompile(regString)

		for i, indicator := range document.Indicators {
			document.Indicators[i].PromQL = reg.ReplaceAllString(indicator.PromQL, v)
		}
	}
}
