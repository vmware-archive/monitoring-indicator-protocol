package indicator

import (
	"github.com/krishicks/yaml-patch"
	"time"
)

type OperatorType int

const (
	LessThan OperatorType = iota
	LessThanOrEqualTo
	EqualTo
	NotEqualTo
	GreaterThanOrEqualTo
	GreaterThan
)

type Document struct {
	APIVersion string
	Product    Product
	Metadata   map[string]string
	Indicators []Indicator
	Layout     Layout
}

type Patch struct {
	APIVersion string
	Match      Match
	Operations []yamlpatch.Operation
}

type Match struct {
	Name     *string
	Version  *string
	Metadata map[string]string
}

type Product struct {
	Name    string
	Version string
}

type Indicator struct {
	Name          string
	PromQL        string
	Thresholds    []Threshold
	Documentation map[string]string
	Presentation  *Presentation
}

type Threshold struct {
	Level    string
	Operator OperatorType
	Value    float64
}

type Presentation struct {
	ChartType
	CurrentValue bool
	Frequency    time.Duration
}

const (
	LineChart ChartType = "line"
	AreaChart ChartType = "area"
	BarChart  ChartType = "bar"
)

type ChartType string

func (e *Threshold) GetComparatorAbbrev() string {
	switch e.Operator {
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

func (e *Threshold) GetComparator() string {
	switch e.Operator {
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
	Title       string
	Description string
	Sections    []Section
	Owner       string
}

type Section struct {
	Title       string
	Description string
	Indicators  []Indicator
}
