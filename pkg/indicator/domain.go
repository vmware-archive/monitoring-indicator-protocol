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

type Documentation struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Sections    []Section `yaml:"sections"`
}

type Section struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Indicators  []string `yaml:"indicators"`
	Metrics     []string `yaml:"metrics"`
}
