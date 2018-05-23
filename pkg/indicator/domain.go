package indicator

type OperatorType int

const (
	LessThan             OperatorType = iota
	LessThanOrEqualTo
	EqualTo
	NotEqualTo
	GreaterThanOrEqualTo
	GreaterThan
)

type Document struct {
	Metrics    []Metric
	Indicators []Indicator
}

type Indicator struct {
	Name        string
	Description string
	PromQL      string
	Thresholds  []Threshold

	Metrics     []string
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
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}
