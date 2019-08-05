package registry

import (
	"time"

	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIDocumentResponse struct {
	APIVersion string                  `json:"apiVersion"`
	UID        string                  `json:"uid"`
	Kind       string                  `json:"kind"`
	Metadata   APIMetadataResponse     `json:"metadata"`
	Spec       APIDocumentSpecResponse `json:"spec"`
}

type APIMetadataResponse struct {
	Labels map[string]string `json:"labels"`
}

type APIDocumentSpecResponse struct {
	Product    APIProductResponse     `json:"product"`
	Indicators []APIIndicatorResponse `json:"indicators"`
	Layout     APILayoutResponse      `json:"layout"`
}

type APIProductResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type APIThresholdResponse struct {
	Level    string  `json:"level"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

type APIPresentationResponse struct {
	ChartType    string   `json:"chartType"`
	CurrentValue bool     `json:"currentValue"`
	Frequency    int64    `json:"frequency"`
	Labels       []string `json:"labels"`
	Units        string   `json:"units"`
}

type APIIndicatorResponse struct {
	Name          string                      `json:"name"`
	PromQL        string                      `json:"promql"`
	Thresholds    []APIThresholdResponse      `json:"thresholds"`
	Alert         APIAlertResponse            `json:"alert"`
	Documentation map[string]string           `json:"documentation,omitempty"`
	Presentation  APIPresentationResponse     `json:"presentation"`
	Status        *APIIndicatorStatusResponse `json:"status"`
}


type APIIndicatorStatusResponse struct {
	Value     *string   `json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type APIAlertResponse struct {
	For  string `json:"for"`
	Step string `json:"step"`
}

type APILayoutResponse struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Sections    []APISectionResponse `json:"sections"`
	Owner       string               `json:"owner"`
}

type APISectionResponse struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Indicators  []string `json:"indicators"`
}

func ToIndicatorDocument(d APIDocumentResponse) v1.IndicatorDocument {
	indicators := make([]v1.IndicatorSpec, 0)
	for _, i := range d.Spec.Indicators {
		indicators = append(indicators, convertIndicator(i))
	}

	return v1.IndicatorDocument{
		TypeMeta: metav1.TypeMeta{
			APIVersion: d.APIVersion,
			Kind:       "IndicatorDocument",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:    types.UID(d.UID),
			Labels: d.Metadata.Labels,
		},

		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{
				Name:    d.Spec.Product.Name,
				Version: d.Spec.Product.Version,
			},
			Indicators: indicators,
			Layout:     convertLayout(d.Spec.Layout),
		},
	}
}

func convertIndicator(i APIIndicatorResponse) v1.IndicatorSpec {
	apiThresholds := i.Thresholds
	thresholds := ConvertThresholds(apiThresholds)

	return v1.IndicatorSpec{
		Name:   i.Name,
		PromQL: i.PromQL,
		Alert: v1.Alert{
			For:  i.Alert.For,
			Step: i.Alert.Step,
		},
		Thresholds:    thresholds,
		Documentation: i.Documentation,
		Presentation: v1.Presentation{
			ChartType:    v1.ChartType(i.Presentation.ChartType),
			CurrentValue: i.Presentation.CurrentValue,
			Frequency:    i.Presentation.Frequency,
			Labels:       i.Presentation.Labels,
		},
	}
}

func ConvertThresholds(apiv0Thresholds []APIThresholdResponse) []v1.Threshold {
	thresholds := make([]v1.Threshold, 0)
	for _, t := range apiv0Thresholds {
		thresholds = append(thresholds, convertThreshold(t))
	}
	return thresholds
}

func convertThreshold(t APIThresholdResponse) v1.Threshold {
	return v1.Threshold{
		Level:    t.Level,
		Operator: v1.GetComparatorFromString(t.Operator),
		Value:    t.Value,
	}
}

func convertLayout(l APILayoutResponse) v1.Layout {
	return v1.Layout{
		Title:       l.Title,
		Description: l.Description,
		Sections:    convertLayoutSections(l.Sections),
		Owner:       l.Owner,
	}
}

func convertLayoutSections(sections []APISectionResponse) []v1.Section {
	apiSections := make([]v1.Section, 0)

	for _, s := range sections {
		apiSections = append(apiSections, convertLayoutSection(s))
	}

	return apiSections
}

func convertLayoutSection(s APISectionResponse) v1.Section {
	return v1.Section{
		Title:       s.Title,
		Description: s.Description,
		Indicators:  s.Indicators,
	}
}

func ToAPIDocumentResponse(doc v1.IndicatorDocument) APIDocumentResponse {
	indicators := make([]APIIndicatorResponse, 0)

	for _, i := range doc.Spec.Indicators {
		thresholds := make([]APIThresholdResponse, 0)
		for _, t := range i.Thresholds {
			thresholds = append(thresholds, APIThresholdResponse{
				Level:    t.Level,
				Operator: v1.GetComparatorAbbrev(t.Operator),
				Value:    t.Value,
			})
		}
		labels := make([]string, 0)
		for _, l := range i.Presentation.Labels {
			labels = append(labels, l)
		}
		presentation := APIPresentationResponse{
			ChartType:    string(i.Presentation.ChartType),
			CurrentValue: i.Presentation.CurrentValue,
			Frequency:    i.Presentation.Frequency,
			Labels:       labels,
			Units:        i.Presentation.Units,
		}

		alert := APIAlertResponse{
			For:  i.Alert.For,
			Step: i.Alert.Step,
		}

		indicators = append(indicators, APIIndicatorResponse{
			Name:          i.Name,
			PromQL:        i.PromQL,
			Thresholds:    thresholds,
			Alert:         alert,
			Documentation: i.Documentation,
			Presentation:  presentation,
			Status:        getStatus(doc, i),
		})
	}

	sections := make([]APISectionResponse, 0)

	for _, s := range doc.Spec.Layout.Sections {
		sections = append(sections, APISectionResponse{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  s.Indicators,
		})
	}

	return APIDocumentResponse{
		APIVersion: doc.APIVersion,
		UID:        doc.BoshUID(),
		Kind:       doc.Kind,
		Metadata: APIMetadataResponse{
			Labels: doc.Labels,
		},
		Spec: APIDocumentSpecResponse{
			Product: APIProductResponse{
				Name:    doc.Spec.Product.Name,
				Version: doc.Spec.Product.Version,
			},
			Indicators: indicators,
			Layout: APILayoutResponse{
				Title:       doc.Spec.Layout.Title,
				Description: doc.Spec.Layout.Description,
				Sections:    sections,
				Owner:       doc.Spec.Layout.Owner,
			},
		},
	}
}

func getStatus(doc v1.IndicatorDocument, i v1.IndicatorSpec) *APIIndicatorStatusResponse {
	status, ok := doc.Status[i.Name]
	if !ok {
		return nil
	}
	return &APIIndicatorStatusResponse{Value: &status.Phase}
}
