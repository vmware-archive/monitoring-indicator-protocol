package indicator

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v2"
)

type yamlDocument struct {
	Metrics       []Metric        `yaml:"metrics"`
	Indicators    []yamlIndicator `yaml:"indicators"`
	Documentation `yaml:"documentation"`
}

type yamlIndicator struct {
	Name        string          `yaml:"name"`
	Title       string          `yaml:"title"`
	Description string          `yaml:"description"`
	Promql      string          `yaml:"promql"`
	Thresholds  []yamlThreshold `yaml:"thresholds"`
	MetricRefs  []MetricRef     `yaml:"metrics"`
	Response    string          `yaml:"response"`
	Measurement string          `yaml:"measurement"`
}

type yamlThreshold struct {
	Level   string `yaml:"level"`
	Dynamic bool   `yaml:"dynamic"`
	LT      string `yaml:"lt"`
	LTE     string `yaml:"lte"`
	EQ      string `yaml:"eq"`
	NEQ     string `yaml:"neq"`
	GTE     string `yaml:"gte"`
	GT      string `yaml:"gt"`
}

func ReadIndicatorDocument(yamlBytes []byte) (Document, error) {
	var d yamlDocument

	err := yaml.Unmarshal(yamlBytes, &d)
	if err != nil {
		return Document{}, fmt.Errorf("could not unmarshal KPIs: %s", err)
	}

	var indicators []Indicator
	for _, yamlKPI := range d.Indicators {
		var thresholds []Threshold
		for _, yamlThreshold := range yamlKPI.Thresholds {
			threshold, err := thresholdFromYAML(yamlThreshold)
			if err != nil {
				return Document{}, fmt.Errorf("could not convert yaml indicators to indicators: %s", err)
			}

			thresholds = append(thresholds, threshold)
		}

		indicators = append(indicators, Indicator{
			Name:        yamlKPI.Name,
			Title:       yamlKPI.Title,
			Description: yamlKPI.Description,
			PromQL:      yamlKPI.Promql,
			Thresholds:  thresholds,
			MetricRefs:  yamlKPI.MetricRefs,
			Response:    yamlKPI.Response,
			Measurement: yamlKPI.Measurement,
		})
	}

	return Document{
		Indicators:    indicators,
		Metrics:       d.Metrics,
		Documentation: d.Documentation,
	}, nil
}

func thresholdFromYAML(threshold yamlThreshold) (Threshold, error) {
	var operator OperatorType
	var value float64
	var err error

	switch {
	case threshold.LT != "":
		operator = LessThan
		value, err = strconv.ParseFloat(threshold.LT, 64)
	case threshold.LTE != "":
		operator = LessThanOrEqualTo
		value, err = strconv.ParseFloat(threshold.LTE, 64)
	case threshold.EQ != "":
		operator = EqualTo
		value, err = strconv.ParseFloat(threshold.EQ, 64)
	case threshold.NEQ != "":
		operator = NotEqualTo
		value, err = strconv.ParseFloat(threshold.NEQ, 64)
	case threshold.GTE != "":
		operator = GreaterThanOrEqualTo
		value, err = strconv.ParseFloat(threshold.GTE, 64)
	case threshold.GT != "":
		operator = GreaterThan
		value, err = strconv.ParseFloat(threshold.GT, 64)
	default:
		return Threshold{}, fmt.Errorf("could not find threshold value: one of [lt, lte, eq, neq, gte, gt] must be provided as a float")
	}

	if err != nil {
		return Threshold{}, err
	}

	return Threshold{
		Level:    threshold.Level,
		Operator: operator,
		Value:    value,
		Dynamic:  threshold.Dynamic,
	}, nil
}
