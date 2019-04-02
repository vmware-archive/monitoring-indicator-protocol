package registry

import (
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

type APIV0Document struct {
	APIVersion string            `json:"apiVersion"`
	Product    APIV0Product      `json:"product"`
	Metadata   map[string]string `json:"metadata"`
	Indicators []APIV0Indicator  `json:"indicators"`
	Layout     APIV0Layout       `json:"layout"`
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

type APIV0Presentation struct {
	ChartType    string   `json:"chartType"`
	CurrentValue bool     `json:"currentValue"`
	Frequency    float64  `json:"frequency"`
	Labels       []string `json:"labels"`
	Units        string   `json:"units"`
}

type APIV0Indicator struct {
	Name          string             `json:"name"`
	PromQL        string             `json:"promql"`
	Thresholds    []APIV0Threshold   `json:"thresholds"`
	Alert         APIV0Alert         `json:"alert"`
	Documentation map[string]string  `json:"documentation,omitempty"`
	Presentation  *APIV0Presentation `json:"presentation"`
}

type APIV0Alert struct {
	For  string `json:"for"`
	Step string `json:"step"`
}

type APIV0Layout struct {
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

func ToIndicatorDocument(d APIV0Document) indicator.Document {
	indicators := make([]indicator.Indicator, 0)
	for _, i := range d.Indicators {
		indicators = append(indicators, convertIndicator(i))
	}

	return indicator.Document{
		APIVersion: d.APIVersion,
		Product: indicator.Product{
			Name:    d.Product.Name,
			Version: d.Product.Version,
		},
		Metadata:   d.Metadata,
		Indicators: indicators,
		Layout:     convertLayout(d.Layout, indicators),
	}
}

func convertIndicator(i APIV0Indicator) indicator.Indicator {
	thresholds := make([]indicator.Threshold, 0)
	for _, t := range i.Thresholds {
		thresholds = append(thresholds, convertThreshold(t))
	}

	return indicator.Indicator{
		Name:       i.Name,
		PromQL:     i.PromQL,
		Thresholds: thresholds,
		Alert: indicator.Alert{
			For:  i.Alert.For,
			Step: i.Alert.Step,
		},
		Documentation: i.Documentation,
		Presentation: &indicator.Presentation{
			ChartType:    indicator.ChartType(i.Presentation.ChartType),
			CurrentValue: i.Presentation.CurrentValue,
			Frequency:    time.Duration(i.Presentation.Frequency),
			Labels:       i.Presentation.Labels,
		},
	}
}

func convertThreshold(t APIV0Threshold) indicator.Threshold {
	return indicator.Threshold{
		Level:    t.Level,
		Operator: indicator.GetComparatorFromString(t.Operator),
		Value:    t.Value,
	}
}

func convertLayout(l APIV0Layout, indicators []indicator.Indicator) indicator.Layout {
	return indicator.Layout{
		Title:       l.Title,
		Description: l.Description,
		Sections:    convertLayoutSections(l.Sections, indicators),
		Owner:       l.Owner,
	}
}

func convertLayoutSections(sections []APIV0Section, indicators []indicator.Indicator) []indicator.Section {
	apiSections := make([]indicator.Section, 0)

	for _, s := range sections {
		apiSections = append(apiSections, convertLayoutSection(s, indicators))
	}

	return apiSections
}

func convertLayoutSection(s APIV0Section, indicators []indicator.Indicator) indicator.Section {
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
		var presentation *APIV0Presentation
		if i.Presentation != nil {
			labels := make([]string, 0)
			for _, l := range i.Presentation.Labels {
				labels = append(labels, l)
			}
			presentation = &APIV0Presentation{
				ChartType:    string(i.Presentation.ChartType),
				CurrentValue: i.Presentation.CurrentValue,
				Frequency:    i.Presentation.Frequency.Seconds(),
				Labels:       labels,
				Units:        i.Presentation.Units,
			}
		}

		alert := APIV0Alert{
			For:  i.Alert.For,
			Step: i.Alert.Step,
		}

		indicators = append(indicators, APIV0Indicator{
			Name:          i.Name,
			PromQL:        i.PromQL,
			Thresholds:    thresholds,
			Alert:         alert,
			Documentation: i.Documentation,
			Presentation:  presentation,
		})
	}

	sections := make([]APIV0Section, 0)

	for _, s := range doc.Layout.Sections {
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
		Layout: APIV0Layout{
			Title:       doc.Layout.Title,
			Description: doc.Layout.Description,
			Sections:    sections,
			Owner:       doc.Layout.Owner,
		},
	}
}
