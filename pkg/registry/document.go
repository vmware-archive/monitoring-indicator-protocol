package registry

import (
	"time"

	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: remove this (once we don't need it...)

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

func ToIndicatorDocument(d APIV0Document) v1alpha1.IndicatorDocument {
	indicators := make([]v1alpha1.IndicatorSpec, 0)
	for _, i := range d.Indicators {
		indicators = append(indicators, convertIndicator(i))
	}

	return v1alpha1.IndicatorDocument{
		TypeMeta: metav1.TypeMeta{
			APIVersion: d.APIVersion,
			Kind:       "IndicatorDocument",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:    types.UID(d.UID),
			Labels: d.Metadata,
		},

		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{
				Name:    d.Product.Name,
				Version: d.Product.Version,
			},
			Indicators: indicators,
			Layout:     convertLayout(d.Layout),
		},
	}
}

func convertIndicator(i APIV0Indicator) v1alpha1.IndicatorSpec {
	apiv0Thresholds := i.Thresholds
	thresholds := ConvertThresholds(apiv0Thresholds)

	return v1alpha1.IndicatorSpec{
		Name:   i.Name,
		PromQL: i.PromQL,
		Alert: v1alpha1.Alert{
			For:  i.Alert.For,
			Step: i.Alert.Step,
		},
		Thresholds: thresholds,
		//ServiceLevel: &v1alpha1.ServiceLevel{
		//	Objective: i.ServiceLevel.Objective,
		//},
		Documentation: i.Documentation,
		Presentation: v1alpha1.Presentation{
			ChartType:    v1alpha1.ChartType(i.Presentation.ChartType),
			CurrentValue: i.Presentation.CurrentValue,
			Frequency:    i.Presentation.Frequency,
			Labels:       i.Presentation.Labels,
		},
	}
}

func ConvertThresholds(apiv0Thresholds []APIV0Threshold) []v1alpha1.Threshold {
	thresholds := make([]v1alpha1.Threshold, 0)
	for _, t := range apiv0Thresholds {
		thresholds = append(thresholds, convertThreshold(t))
	}
	return thresholds
}

func convertThreshold(t APIV0Threshold) v1alpha1.Threshold {
	return v1alpha1.Threshold{
		Level:    t.Level,
		Operator: indicator.GetComparatorFromString(t.Operator),
		Value:    t.Value,
	}
}

func convertLayout(l APIV0Layout) v1alpha1.Layout {
	return v1alpha1.Layout{
		Title:       l.Title,
		Description: l.Description,
		Sections:    convertLayoutSections(l.Sections),
		Owner:       l.Owner,
	}
}

func convertLayoutSections(sections []APIV0Section) []v1alpha1.Section {
	apiSections := make([]v1alpha1.Section, 0)

	for _, s := range sections {
		apiSections = append(apiSections, convertLayoutSection(s))
	}

	return apiSections
}

func convertLayoutSection(s APIV0Section) v1alpha1.Section {
	return v1alpha1.Section{
		Title:       s.Title,
		Description: s.Description,
		Indicators:  s.Indicators,
	}
}

func ToAPIV0Document(doc v1alpha1.IndicatorDocument) APIV0Document {
	indicators := make([]APIV0Indicator, 0)

	for _, i := range doc.Spec.Indicators {
		thresholds := make([]APIV0Threshold, 0)
		for _, t := range i.Thresholds {
			thresholds = append(thresholds, APIV0Threshold{
				Level:    t.Level,
				Operator: indicator.GetComparatorAbbrev(t.Operator),
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
			Status:        getStatus(doc, i),
		})
	}

	sections := make([]APIV0Section, 0)

	for _, s := range doc.Spec.Layout.Sections {
		sections = append(sections, APIV0Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  s.Indicators,
		})
	}

	return APIV0Document{
		APIVersion: doc.APIVersion,
		UID:        doc.BoshUID(),
		Product: APIV0Product{
			Name:    doc.Spec.Product.Name,
			Version: doc.Spec.Product.Version,
		},
		Metadata:   doc.ObjectMeta.Labels,
		Indicators: indicators,
		Layout: APIV0Layout{
			Title:       doc.Spec.Layout.Title,
			Description: doc.Spec.Layout.Description,
			Sections:    sections,
			Owner:       doc.Spec.Layout.Owner,
		},
	}
}

func getStatus(doc v1alpha1.IndicatorDocument, i v1alpha1.IndicatorSpec) *APIV0IndicatorStatus {
	status, ok := doc.Status[i.Name]
	if !ok {
		return nil
	}
	return &APIV0IndicatorStatus{Value: &status.Phase}
}

func convertServiceLevel(level *v1alpha1.ServiceLevel) *APIV0ServiceLevel {
	if level == nil {
		return nil
	}
	return &APIV0ServiceLevel{
		Objective: level.Objective,
	}
}
