package v1alpha1

import (
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Product    Product         `json:"product"`
	Indicators []IndicatorSpec `json:"indicators"`
	Layout     Layout          `json:"layout"`
}

type Product struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Indicator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IndicatorSpec
}

type IndicatorSpec struct {
	Product       string            `json:"product"`
	Name          string            `json:"name"`
	Promql        string            `json:"promql"`
	Alert         Alert             `json:"alert"`
	Thresholds    []Threshold       `json:"thresholds"`
	Documentation map[string]string `json:"documentation,omitempty"`
	Presentation  Presentation      `json:"presentation"`
}

type Presentation struct {
	ChartType    indicator.ChartType `json:"chartType"`
	CurrentValue bool                `json:"currentValue"`
	Frequency    int64               `json:"frequency"`
	Labels       []string            `json:"labels"`
}

type Alert struct {
	For  string `json:"for"`
	Step string `json:"step"`
}

type Threshold struct {
	Level string   `json:"level,omitempty"`
	Lt    *float64 `json:"lt,omitempty"`
	Lte   *float64 `json:"lte,omitempty"`
	Eq    *float64 `json:"eq,omitempty"`
	Neq   *float64 `json:"neq,omitempty"`
	Gte   *float64 `json:"gte,omitempty"`
	Gt    *float64 `json:"gt,omitempty"`
}

type Layout struct {
	Owner       string    `json:"owner"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Sections    []Section `json:"sections"`
}

type Section struct {
	Title       string   `json:"title"`
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IndicatorList is a list of Indicator resources
type IndicatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Indicator `json:"items"`
}
