package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
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

	Spec   IndicatorSpec   `json:"spec"`
	Status IndicatorStatus `json:"status"`
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

type IndicatorStatus struct {
	Phase string `json:"phase"`
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
	Level    string  `json:"level,omitempty"`
	Operator string  `json:"operator,omitempty"`
	Value    float64 `json:"value,omitempty"`
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
