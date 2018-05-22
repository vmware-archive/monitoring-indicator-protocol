package kpi

type OperatorType int

const (
	LessThan OperatorType = iota
	LessThanOrEqualTo
	EqualTo
	NotEqualTo
	GreaterThanOrEqualTo
	GreaterThan
)

type KPI struct {
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
	Operator OperatorType
	Value    float64
}

type Event struct {
	Tags map[string]string
	Value float64
	ThresholdLevel string
	ThresholdValue float64
}

//map[string]interface{}
