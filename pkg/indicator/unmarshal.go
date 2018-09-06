package indicator

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v2"
)

func ReadIndicatorDocument(yamlBytes []byte) (Document, error) {
	var d yamlDocument

	err := yaml.Unmarshal(yamlBytes, &d)
	if err != nil {
		return Document{}, fmt.Errorf("could not unmarshal indicators: %s", err)
	}

	var indicators []Indicator
	for idx, yamlKPI := range d.Indicators {
		var thresholds []Threshold
		for _, yamlThreshold := range yamlKPI.Thresholds {
			threshold, err := thresholdFromYAML(yamlThreshold)
			if err != nil {
				return Document{}, fmt.Errorf("could not convert yaml indicators to indicators: %s", err)
			}

			thresholds = append(thresholds, threshold)
		}

		var metrics []Metric
		for mIdx, m := range yamlKPI.MetricRefs {
			metric, ok := findMetric(m, d.Metrics)
			if !ok {
				return Document{}, fmt.Errorf("indicators[%d].metrics[%d] references non-existent metric", idx, mIdx)
			}

			metrics = append(metrics, metric)
		}

		indicators = append(indicators, Indicator{
			Name:        yamlKPI.Name,
			Title:       yamlKPI.Title,
			Description: yamlKPI.Description,
			PromQL:      yamlKPI.Promql,
			Thresholds:  thresholds,
			Metrics:     metrics,
			Response:    yamlKPI.Response,
			Measurement: yamlKPI.Measurement,
		})
	}

	var sections []Section
	for idx, s := range d.YAMLDoc.Sections {

		var sectionMetrics []Metric
		for mIdx, m := range s.MetricRefs {
			metric, ok := findMetric(m, d.Metrics)
			if !ok {
				return Document{}, fmt.Errorf("documentation.sections[%d].metrics[%d] references non-existent metric", idx, mIdx)
			}

			sectionMetrics = append(sectionMetrics, metric)
		}

		var sectionIndicators []Indicator
		for iIdx, i := range s.IndicatorRefs {
			indic, ok := findIndicator(i, indicators)
			if !ok {
				return Document{}, fmt.Errorf("documentation.sections[%d].indicators[%d] references non-existent indicator", idx, iIdx)
			}

			sectionIndicators = append(sectionIndicators, indic)
		}

		sections = append(sections, Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  sectionIndicators,
			Metrics:     sectionMetrics,
		})
	}

	documentation := Documentation{
		Title:       d.YAMLDoc.Title,
		Description: d.YAMLDoc.Description,
		Sections:    sections,
		Owner:       d.YAMLDoc.Owner,
	}

	return Document{
		Labels:        d.Labels,
		Indicators:    indicators,
		Metrics:       d.Metrics,
		Documentation: documentation,
	}, nil
}

type yamlDocument struct {
	Labels     map[string]string `yaml:"labels"`
	Metrics    []Metric          `yaml:"metrics"`
	Indicators []yamlIndicator   `yaml:"indicators"`
	YAMLDoc    yamlDocumentation `yaml:"documentation"`
}

type yamlDocumentation struct {
	Title       string        `yaml:"title"`
	Description string        `yaml:"description"`
	Sections    []yamlSection `yaml:"sections"`
	Owner       string        `yaml:"owner"`
}

type yamlSection struct {
	Title         string         `yaml:"title"`
	Description   string         `yaml:"description"`
	IndicatorRefs []indicatorRef `yaml:"indicators"`
	MetricRefs    []metricRef    `yaml:"metrics"`
}

type yamlIndicator struct {
	Name        string          `yaml:"name"`
	Title       string          `yaml:"title"`
	Description string          `yaml:"description"`
	Promql      string          `yaml:"promql"`
	Thresholds  []yamlThreshold `yaml:"thresholds"`
	MetricRefs  []metricRef     `yaml:"metrics"`
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

type indicatorRef struct {
	Name string `yaml:"name"`
}

func findIndicator(reference indicatorRef, indicators []Indicator) (Indicator, bool) {
	for _, i := range indicators {
		if i.Name == reference.Name {
			return i, true
		}
	}

	return Indicator{}, false
}

type metricRef struct {
	Name     string `yaml:"name"`
	SourceID string `yaml:"source_id"`
}

func findMetric(reference metricRef, metrics []Metric) (Metric, bool) {
	for _, m := range metrics {
		if m.Name == reference.Name && m.SourceID == reference.SourceID {
			return m, true
		}
	}

	return Metric{}, false
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
