package registry

import "code.cloudfoundry.org/indicators/pkg/indicator"

type APIV0Document struct {
	APIVersion    string             `json:"apiVersion"`
	Product       APIV0Product       `json:"product"`
	Metadata      map[string]string  `json:"metadata"`
	Indicators    []APIV0Indicator   `json:"indicators"`
	Documentation APIV0Documentation `json:"documentation"`
}

type APIV0Product struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type APIV0Threshold struct {
	Level    string  `json:"level"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

type APIV0Indicator struct {
	Name          string            `json:"name"`
	PromQL        string            `json:"promql"`
	Thresholds    []APIV0Threshold  `json:"thresholds"`
	Documentation map[string]string `json:"documentation"`
	SLO           float64           `json:"slo,omitempty"`
}

type APIV0Documentation struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Sections    []APIV0Section `json:"sections"`
	Owner       string         `json:"owner"`
}

type APIV0Section struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Indicators  []string `json:"indicators"`
}

func ToAPIV0Document(doc indicator.Document) APIV0Document {
	indicators := make([]APIV0Indicator, 0)

	for _, i := range doc.Indicators {
		thresholds := make([]APIV0Threshold, 0)
		for _, t := range i.Thresholds {
			thresholds = append(thresholds, APIV0Threshold{
				Level:    t.Level,
				Operator: t.GetComparatorAbbrev(),
				Value:    t.Value,
			})
		}

		indicators = append(indicators, APIV0Indicator{
			Name:          i.Name,
			PromQL:        i.PromQL,
			Thresholds:    thresholds,
			SLO:           i.SLO,
			Documentation: i.Documentation,
		})
	}

	sections := make([]APIV0Section, 0)

	for _, s := range doc.Documentation.Sections {
		indicatorNames := make([]string, 0)
		for _, i := range s.Indicators {
			indicatorNames = append(indicatorNames, i.Name)
		}

		sections = append(sections, APIV0Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  indicatorNames,
		})
	}

	return APIV0Document{
		APIVersion: doc.APIVersion,
		Product: APIV0Product{
			Name:    doc.Product.Name,
			Version: doc.Product.Version,
		},
		Metadata:   doc.Metadata,
		Indicators: indicators,
		Documentation: APIV0Documentation{
			Title:       doc.Documentation.Title,
			Description: doc.Documentation.Description,
			Sections:    nil,
			Owner:       doc.Documentation.Owner,
		},
	}
}
