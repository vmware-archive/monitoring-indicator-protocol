package registry

import (
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

type APIV0Document struct {
	APIVersion string            `json:"apiVersion"`
	UID        string            `json:"uid"`
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
	Frequency    int64    `json:"frequency"`
	Labels       []string `json:"labels"`
	Units        string   `json:"units"`
}

type APIV0Indicator struct {
	Name          string                `json:"name"`
	PromQL        string                `json:"promql"`
	Thresholds    []APIV0Threshold      `json:"thresholds"`
	Alert         APIV0Alert            `json:"alert"`
	ServiceLevel  *APIV0ServiceLevel    `json:"serviceLevel"`
	Documentation map[string]string     `json:"documentation,omitempty"`
	Presentation  APIV0Presentation     `json:"presentation"`
	Status        *APIV0IndicatorStatus `json:"status"`
}

type APIV0ServiceLevel struct {
	Objective float64 `json:"objective"`
}

type APIV0IndicatorStatus struct {
	Value     *string   `json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
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
		Layout:     convertLayout(d.Layout),
	}
}

func convertIndicator(i APIV0Indicator) indicator.Indicator {
	apiv0Thresholds := i.Thresholds
	thresholds := ConvertThresholds(apiv0Thresholds)

	return indicator.Indicator{
		Name:       i.Name,
		PromQL:     i.PromQL,
		Thresholds: thresholds,
		Alert: indicator.Alert{
			For:  i.Alert.For,
			Step: i.Alert.Step,
		},
		Documentation: i.Documentation,
		Presentation: indicator.Presentation{
			ChartType:    indicator.ChartType(i.Presentation.ChartType),
			CurrentValue: i.Presentation.CurrentValue,
			Frequency:    i.Presentation.Frequency,
			Labels:       i.Presentation.Labels,
		},
	}
}

func ConvertThresholds(apiv0Thresholds []APIV0Threshold) []indicator.Threshold {
	thresholds := make([]indicator.Threshold, 0)
	for _, t := range apiv0Thresholds {
		thresholds = append(thresholds, convertThreshold(t))
	}
	return thresholds
}

func convertThreshold(t APIV0Threshold) indicator.Threshold {
	return indicator.Threshold{
		Level:    t.Level,
		Operator: indicator.GetComparatorFromString(t.Operator),
		Value:    t.Value,
	}
}

func convertLayout(l APIV0Layout) indicator.Layout {
	return indicator.Layout{
		Title:       l.Title,
		Description: l.Description,
		Sections:    convertLayoutSections(l.Sections),
		Owner:       l.Owner,
	}
}

func convertLayoutSections(sections []APIV0Section) []indicator.Section {
	apiSections := make([]indicator.Section, 0)

	for _, s := range sections {
		apiSections = append(apiSections, convertLayoutSection(s))
	}

	return apiSections
}

func convertLayoutSection(s APIV0Section) indicator.Section {
	return indicator.Section{
		Title:       s.Title,
		Description: s.Description,
		Indicators:  s.Indicators,
	}
}

func ToAPIV0Document(doc indicator.Document, getStatus func(string) *APIV0IndicatorStatus) APIV0Document {
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
		labels := make([]string, 0)
		for _, l := range i.Presentation.Labels {
			labels = append(labels, l)
		}
		presentation := APIV0Presentation{
			ChartType:    string(i.Presentation.ChartType),
			CurrentValue: i.Presentation.CurrentValue,
			Frequency:    i.Presentation.Frequency,
			Labels:       labels,
			Units:        i.Presentation.Units,
		}

		alert := APIV0Alert{
			For:  i.Alert.For,
			Step: i.Alert.Step,
		}
		serviceLevel := convertServiceLevel(i.ServiceLevel)

		indicators = append(indicators, APIV0Indicator{
			Name:          i.Name,
			PromQL:        i.PromQL,
			Thresholds:    thresholds,
			Alert:         alert,
			ServiceLevel:  serviceLevel,
			Documentation: i.Documentation,
			Presentation:  presentation,
			Status:        getStatus(i.Name),
		})
	}

	sections := make([]APIV0Section, 0)

	for _, s := range doc.Layout.Sections {
		sections = append(sections, APIV0Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  s.Indicators,
		})
	}

	return APIV0Document{
		APIVersion: doc.APIVersion,
		UID:        doc.UID(),
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

func convertServiceLevel(level *indicator.ServiceLevel) *APIV0ServiceLevel {
	if level == nil {
		return nil
	}
	return &APIV0ServiceLevel{
		Objective: level.Objective,
	}
}
