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

	MetricRefs  []MetricRef
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
	Description string `yaml:"description"`
}

type Documentation struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Sections    []Section `yaml:"sections"`
	Owner       string    `yaml:"owner"`
}

type Section struct {
	Title         string         `yaml:"title"`
	Description   string         `yaml:"description"`
	IndicatorRefs []IndicatorRef `yaml:"indicators"`
	MetricRefs    []MetricRef    `yaml:"metrics"`
}

type IndicatorRef struct {
	Name string `yaml:"name"`
}

func FindIndicator(reference IndicatorRef, indicators []Indicator) (Indicator, bool) {
	for _, i := range indicators {
		if i.Name == reference.Name {
			return i, true
		}
	}

	return Indicator{}, false
}

type MetricRef struct {
	Name     string `yaml:"name"`
	SourceID string `yaml:"source_id"`
}

func FindMetric(reference MetricRef, metrics []Metric) (Metric, bool) {
	for _, m := range metrics {
		if m.Name == reference.Name && m.SourceID == reference.SourceID {
			return m, true
		}
	}

	return Metric{}, false
}
