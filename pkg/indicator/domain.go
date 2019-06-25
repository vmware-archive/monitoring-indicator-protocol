package indicator

import (
	"crypto/sha1"
	"fmt"
	"sort"

	"github.com/cppforlife/go-patch/patch"
)

type OperatorType int

const (
	Undefined OperatorType = iota
	LessThan
	LessThanOrEqualTo
	EqualTo
	NotEqualTo
	GreaterThanOrEqualTo
	GreaterThan
)

type Document struct {
	APIVersion string            `yaml:"apiVersion"`
	Product    Product           `yaml:"product"`
	Metadata   map[string]string `yaml:"metadata"`
	Indicators []Indicator       `yaml:"indicators"`
	Layout     Layout            `yaml:"layout"`
}

func (document Document) UID() string {
	return fmt.Sprintf("%s-%x", document.Product.Name, getMetadataSHA(document.Metadata))
}

func (document Document) GetIndicator(name string) (Indicator, bool) {
	for _, indicator := range document.Indicators {
		if indicator.Name == name {
			return indicator, true
		}
	}
	return Indicator{}, false
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

type Patch struct {
	APIVersion string
	Match      Match
	Operations []patch.OpDefinition
}

type Match struct {
	Name     *string
	Version  *string
	Metadata map[string]string
}

type Product struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Indicator struct {
	Name          string            `yaml:"name"`
	PromQL        string            `yaml:"promql"`
	Thresholds    []Threshold       `yaml:"thresholds,omitempty"`
	Alert         Alert             `yaml:"alert"`
	ServiceLevel  *ServiceLevel     `yaml:"serviceLevel"`
	Documentation map[string]string `yaml:"documentation,omitempty"`
	Presentation  Presentation      `yaml:"presentation"`
}

type ServiceLevel struct {
	Objective float64 `yaml:"objective"`
}

type Alert struct {
	For  string `yaml:"for"`
	Step string `yaml:"step"`
}

type Threshold struct {
	Level    string       `yaml:"level"`
	Operator OperatorType `yaml:"operator"`
	Value    float64      `yaml:"value"`
}

type Presentation struct {
	ChartType    ChartType `yaml:"chartType"`
	CurrentValue bool      `yaml:"currentValue"`
	Frequency    int64     `yaml:"frequency"`
	Labels       []string  `yaml:"labels"`
	Units        string    `yaml:"units"`
}

type ChartType string

const (
	StepChart   ChartType = "step"
	BarChart    ChartType = "bar"
	StatusChart ChartType = "status"
	QuotaChart  ChartType = "quota"
)

var ChartTypes = []ChartType{StepChart, BarChart, StatusChart, QuotaChart}

func (t *Threshold) GetComparatorAbbrev() string {
	switch t.Operator {
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
func GetComparatorFromString(operator string) OperatorType {
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
		return -1
	}
}

func (t *Threshold) GetComparator() string {
	switch t.Operator {
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
	}

	return ""
}

type Layout struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Sections    []Section `yaml:"sections"`
	Owner       string    `yaml:"owner"`
}

type Section struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Indicators  []string `yaml:"indicators"`
}
