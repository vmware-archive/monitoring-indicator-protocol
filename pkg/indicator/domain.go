package indicator

type OperatorType int

const (
	LessThan OperatorType = iota
	LessThanOrEqualTo
	EqualTo
	NotEqualTo
	GreaterThanOrEqualTo
	GreaterThan
)

func GetComparatorDescription(e OperatorType) string {
	switch e {
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

func GetComparator(e OperatorType) string {
	switch e {
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

type Document struct {
	APIVersion    string
	Product       string
	Version       string
	Metadata      map[string]string
	Indicators    []Indicator
	Documentation Documentation
}

type Indicator struct {
	Name          string
	PromQL        string
	Thresholds    []Threshold
	SLO           float64
	Documentation map[string]string
}

type Threshold struct {
	Level    string
	Operator OperatorType
	Value    float64
}

type Documentation struct {
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
