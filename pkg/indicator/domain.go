package indicator

import "github.com/krishicks/yaml-patch"

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
	Origin     string
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
}

type Threshold struct {
	Level    string
	Operator OperatorType
	Value    float64
}

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
