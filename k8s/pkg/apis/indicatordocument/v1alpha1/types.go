package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IndicatorDocument is a specification for a IndicatorDocument resource
type IndicatorDocument struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IndicatorDocumentSpec   `json:"spec"`
	Status IndicatorDocumentStatus `json:"status"`
}

// IndicatorDocumentSpec is the spec for a IndicatorDocument resource
type IndicatorDocumentSpec struct {
	Product    Product     `json:"product"`
	Indicators []Indicator `json:"indicators"`
	Layout     Layout      `json:"layout"`
}

type Product struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Indicator struct {
	Name          string            `json:"name"`
	Promql        string            `json:"promql"`
	Alert         Alert             `json:"alert"`
	Thresholds    []Threshold       `json:"thresholds"`
	Documentation map[string]string `json:"documentation"`
}

type Alert struct {
	For  string `json:"for"`
	Step string `json:"step"`
}

type Threshold struct {
	Level string   `json:"level"`
	Lt    *float64 `json:"lt"`
	Lte   *float64 `json:"lte"`
	Eq    *float64 `json:"eq"`
	Neq   *float64 `json:"neq"`
	Gte   *float64 `json:"gte"`
	Gt    *float64 `json:"gt"`
}

type Layout struct {
	Owner       string    `json:"owner"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Sections    []Section `json:"sections"`
}

type Section struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Indicators  []string `json:"indicators"`
}

// IndicatorDocumentStatus is the status for a IndicatorDocument resource
type IndicatorDocumentStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IndicatorDocumentList is a list of IndicatorDocument resources
type IndicatorDocumentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []IndicatorDocument `json:"items"`
}
