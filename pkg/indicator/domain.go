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

func (e OperatorType) String() string {
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

type Document struct {
	APIVersion    string
	Labels        map[string]string
	Metrics       []Metric
	Indicators    []Indicator
	Documentation Documentation
}

type Indicator struct {
	Name        string
	Title       string
	Description string
	PromQL      string
	Thresholds  []Threshold
	Metrics     []Metric
	Response    string
	Measurement string
}

type Threshold struct {
	Level    string
	Dynamic  bool
	Operator OperatorType
	Value    float64
}

type Metric struct {
	Title       string `yaml:"title"`
	Origin      string `yaml:"origin"`
	SourceID    string `yaml:"source_id"`
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Frequency   string `yaml:"frequency"`
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
	Metrics     []Metric
}
