package indicator

import (
	"time"

	"github.com/cppforlife/go-patch/patch"
)

type OperatorType int

const (
	LessThan OperatorType = iota
	LessThanOrEqualTo
	EqualTo
	NotEqualTo
	GreaterThanOrEqualTo
	GreaterThan
	Undefined
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
	Operations []patch.OpDefinition
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
	Alert         Alert
	Documentation map[string]string
	Presentation  *Presentation
}

type Alert struct {
	For  string
	Step string
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
	Labels       []string
	Units        string
}

type ChartType string

const (
	StepChart ChartType = "step"
	BarChart  ChartType = "bar"
	StatusChart  ChartType = "status"
)

var ChartTypes = []ChartType{StepChart, BarChart, StatusChart}

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
