package v1

import (
	"crypto/sha1"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IndicatorDocument is a specification for a IndicatorDocument resource
type IndicatorDocument struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IndicatorDocumentSpec   `json:"spec"`
	Status IndicatorDocumentStatus `json:"status,omitempty"`
}

func (id IndicatorDocument) BoshUID() string {
	return fmt.Sprintf("%s-%x", id.Spec.Product.Name, getMetadataSHA(id.Labels))
}

func (id IndicatorDocument) Indicator(name string) *IndicatorSpec {
	for _, indicator := range id.Spec.Indicators {
		if indicator.Name == name {
			return &indicator
		}
	}
	return nil
}

func getMetadataSHA(metadata map[string]string) [20]byte {
	var metadataKeys []string
	for k := range metadata {
		metadataKeys = append(metadataKeys, k)
	}
	sort.Strings(metadataKeys)

	metadataKey := ""
	for _, k := range metadataKeys {
		metadataKey = fmt.Sprintf("%s:%s", k, metadata[k])
	}

	return sha1.Sum([]byte(metadataKey))
}

// IndicatorDocumentSpec is the spec for a IndicatorDocument resource
type IndicatorDocumentSpec struct {
	Product    Product         `json:"product"`
	Indicators []IndicatorSpec `json:"indicators,omitempty"`
	Layout     Layout          `json:"layout,omitempty"`
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
	Status IndicatorStatus `json:"status,omitempty"`
}

type IndicatorSpec struct {
	// Product duplicated between here and indicator documents for the `kubectl get indicators` display
	Product       string            `json:"product,omitempty"`
	Name          string            `json:"name"`
	PromQL        string            `json:"promql"`
	Alert         Alert             `json:"alert,omitempty"`
	Thresholds    []Threshold       `json:"thresholds,omitempty"`
	Documentation map[string]string `json:"documentation,omitempty"`
	Presentation  Presentation      `json:"presentation,omitempty"`
}

type IndicatorStatus struct {
	Phase     string      `json:"phase"`
	UpdatedAt metav1.Time `json:"updatedAt"`
}

type Presentation struct {
	ChartType    ChartType `json:"chartType,omitempty"`
	CurrentValue bool      `json:"currentValue,omitempty"`
	Frequency    int64     `json:"frequency,omitempty"`
	Labels       []string  `json:"labels,omitempty"`
	Units        string    `json:"units,omitempty"`
}

type ChartType string

const (
	StepChart   ChartType = "step"
	BarChart    ChartType = "bar"
	StatusChart ChartType = "status"
	QuotaChart  ChartType = "quota"
)

var ChartTypes = []ChartType{StepChart, BarChart, StatusChart, QuotaChart}

func (chartType *ChartType) Validate(idx int) []error {
	var es []error
	valid := false
	for _, validChartType := range ChartTypes {
		if *chartType == validChartType {
			valid = true
		}
	}
	if !valid {
		es = append(es, fmt.Errorf("indicators[%d] invalid chartType provided - valid chart types are %v", idx, ChartTypes))
	}

	return es
}

type Alert struct {
	For  string `json:"for,omitempty"`
	Step string `json:"step,omitempty"`
}

type Threshold struct {
	Level    string            `json:"level"`
	Operator ThresholdOperator `json:"operator"`
	Value    float64           `json:"value"`
}

type ThresholdOperator int

func (ot ThresholdOperator) MarshalJSON() ([]byte, error) {
	operatorString := GetComparatorAbbrev(ot)
	if operatorString == "" {
		return []byte("null"), nil
	}
	quoteWrappedString := fmt.Sprintf(`"%s"`, operatorString)
	return []byte(quoteWrappedString), nil
}

func (ot *ThresholdOperator) UnmarshalJSON(data []byte) error {
	*ot = unmarshalComparatorFromString(string(data))
	if *ot == -1 {
		*ot = Undefined
		return nil
	}
	return nil
}

const (
	Undefined ThresholdOperator = iota
	LessThan
	LessThanOrEqualTo
	EqualTo
	NotEqualTo
	GreaterThanOrEqualTo
	GreaterThan
)

func unmarshalComparatorFromString(operator string) ThresholdOperator {
	switch operator {
	case `"lt"`:
		return LessThan
	case `"lte"`:
		return LessThanOrEqualTo
	case `"eq"`:
		return EqualTo
	case `"neq"`:
		return NotEqualTo
	case `"gte"`:
		return GreaterThanOrEqualTo
	case `"gt"`:
		return GreaterThan
	default:
		return Undefined
	}
}

func GetComparatorFromString(operator string) ThresholdOperator {
	switch operator {
	case "lt":
		return LessThan
	case "lte":
		return LessThanOrEqualTo
	case "eq":
		return EqualTo
	case "neq":
		return NotEqualTo
	case "gte":
		return GreaterThanOrEqualTo
	case "gt":
		return GreaterThan
	default:
		return Undefined
	}
}

func GetComparatorAbbrev(op ThresholdOperator) string {
	switch op {
	case LessThan:
		return "lt"
	case LessThanOrEqualTo:
		return "lte"
	case EqualTo:
		return "eq"
	case NotEqualTo:
		return "neq"
	case GreaterThanOrEqualTo:
		return "gte"
	case GreaterThan:
		return "gt"
	default:
		return ""
	}
}

func GetComparatorSymbol(op ThresholdOperator) string {
	switch op {
	case LessThan:
		return "<"
	case LessThanOrEqualTo:
		return "<="
	case EqualTo:
		return "=="
	case NotEqualTo:
		return "!="
	case GreaterThanOrEqualTo:
		return ">="
	case GreaterThan:
		return ">"
	default:
		return ""
	}
}

type Layout struct {
	Owner       string    `json:"owner,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Sections    []Section `json:"sections,omitempty"`
}

type Section struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Indicators  []string `json:"indicators,omitempty"`
}

// IndicatorDocumentStatus is the status for a IndicatorDocument resource,
// it maps names of indicators to their statuses.
type IndicatorDocumentStatus map[string]IndicatorStatus

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
