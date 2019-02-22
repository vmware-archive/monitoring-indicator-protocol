package registry

import "github.com/pivotal/indicator-protocol/pkg/indicator"

type apiV0Document struct {
	APIVersion string            `json:"apiVersion"`
	Product    apiV0Product      `json:"product"`
	Metadata   map[string]string `json:"metadata"`
	Indicators []apiV0Indicator  `json:"indicators"`
	Layout     apiV0Layout       `json:"layout"`
}

type apiV0Product struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type apiV0Threshold struct {
	Level    string  `json:"level"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

type apiV0Presentation struct {
	ChartType    string   `json:"chartType"`
	CurrentValue bool     `json:"currentValue"`
	Frequency    float64  `json:"frequency"`
	Labels       []string `json:"labels"`
}

type apiV0Indicator struct {
	Name          string             `json:"name"`
	PromQL        string             `json:"promql"`
	Thresholds    []apiV0Threshold   `json:"thresholds,omitempty"`
	Documentation map[string]string  `json:"documentation,omitempty"`
	Presentation  *apiV0Presentation `json:"presentation,omitempty"`
}

type apiV0Layout struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Sections    []apiV0Section `json:"sections,omitempty"`
	Owner       string         `json:"owner"`
}

type apiV0Section struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Indicators  []string `json:"indicators,omitempty"`
}

func toIndicatorDocument(d apiV0Document) indicator.Document {
	indicators := make([]indicator.Indicator, 0)
	for _, i := range d.Indicators {
		indicators = append(indicators, convertIndicator(i))
	}

	return indicator.Document{
		Product: indicator.Product{
			Name:    d.Product.Name,
			Version: d.Product.Version,
		},
		Indicators: indicators,
		Layout:     convertLayout(d.Layout, indicators),
	}
}

func convertIndicator(i apiV0Indicator) indicator.Indicator {
	thresholds := make([]indicator.Threshold, 0)
	for _, t := range i.Thresholds {
		thresholds = append(thresholds, convertThreshold(t))
	}

	return indicator.Indicator{
		Name:          i.Name,
		PromQL:        i.PromQL,
		Thresholds:    thresholds,
		Documentation: i.Documentation,
	}
}

func convertThreshold(t apiV0Threshold) indicator.Threshold {
	return indicator.Threshold{
		Level:    t.Level,
		Operator: indicator.GetComparatorFromString(t.Operator),
		Value:    t.Value,
	}
}

func convertLayout(l apiV0Layout, indicators []indicator.Indicator) indicator.Layout {
	return indicator.Layout{
		Title:       l.Title,
		Description: l.Description,
		Sections:    convertLayoutSections(l.Sections, indicators),
		Owner:       l.Owner,
	}
}

func convertLayoutSections(sections []apiV0Section, indicators []indicator.Indicator) []indicator.Section {
	apiSections := make([]indicator.Section, 0)

	for _, s := range sections {
		apiSections = append(apiSections, convertLayoutSection(s, indicators))
	}

	return apiSections
}

func convertLayoutSection(s apiV0Section, indicators []indicator.Indicator) indicator.Section {
	sectionIndicators := make([]indicator.Indicator, 0)

	for _, name := range s.Indicators {
		for _, i := range indicators {
			if i.Name == name {
				sectionIndicators = append(sectionIndicators, i)
			}
		}
	}

	return indicator.Section{
		Title:       s.Title,
		Description: s.Description,
		Indicators:  sectionIndicators,
	}
}

func toAPIV0Document(doc indicator.Document) apiV0Document {
	indicators := make([]apiV0Indicator, 0)

	for _, i := range doc.Indicators {
		thresholds := make([]apiV0Threshold, 0)
		for _, t := range i.Thresholds {
			thresholds = append(thresholds, apiV0Threshold{
				Level:    t.Level,
				Operator: t.GetComparatorAbbrev(),
				Value:    t.Value,
			})
		}
		var presentation *apiV0Presentation
		if i.Presentation != nil {
			labels := make([]string, 0)
			for _, l := range i.Presentation.Labels {
				labels = append(labels, l)
			}
			presentation = &apiV0Presentation{
				ChartType:    string(i.Presentation.ChartType),
				CurrentValue: i.Presentation.CurrentValue,
				Frequency:    i.Presentation.Frequency.Seconds(),
				Labels:       labels,
			}
		}

		indicators = append(indicators, apiV0Indicator{
			Name:          i.Name,
			PromQL:        i.PromQL,
			Thresholds:    thresholds,
			Documentation: i.Documentation,
			Presentation:  presentation,
		})
	}

	sections := make([]apiV0Section, 0)

	for _, s := range doc.Layout.Sections {
		indicatorNames := make([]string, 0)
		for _, i := range s.Indicators {
			indicatorNames = append(indicatorNames, i.Name)
		}

		sections = append(sections, apiV0Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  indicatorNames,
		})
	}

	return apiV0Document{
		APIVersion: doc.APIVersion,
		Product: apiV0Product{
			Name:    doc.Product.Name,
			Version: doc.Product.Version,
		},
		Metadata:   doc.Metadata,
		Indicators: indicators,
		Layout: apiV0Layout{
			Title:       doc.Layout.Title,
			Description: doc.Layout.Description,
			Sections:    sections,
			Owner:       doc.Layout.Owner,
		},
	}
}
