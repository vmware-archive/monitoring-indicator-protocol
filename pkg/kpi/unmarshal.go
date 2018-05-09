package kpi

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v2"
)

type YamlKPI struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Promql      string          `yaml:"promql"`
	Thresholds  []YamlThreshold `yaml:"thresholds"`
}

type YamlThreshold struct {
	Level string `yaml:"level"`
	LT    string `yaml:"lt"`
	LTE   string `yaml:"lte"`
	EQ    string `yaml:"eq"`
	NEQ   string `yaml:"neq"`
	GTE   string `yaml:"gte"`
	GT    string `yaml:"gt"`
}

func ReadKPIsFromYaml(kpisYAML []byte) ([]KPI, error) {
	yamlTypedKPIs := make([]YamlKPI, 0)

	err := yaml.Unmarshal(kpisYAML, &yamlTypedKPIs)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal KPIs: %s", err)
	}

	var kpis []KPI
	for _, yamlKPI := range yamlTypedKPIs {
		var thresholds []Threshold
		for _, yamlThreshold := range yamlKPI.Thresholds {
			threshold, err := kpiThresholdFromYaml(yamlThreshold)
			if err != nil {
				return nil, fmt.Errorf("could not convert yaml kpis to kpis: %s", err)
			}

			thresholds = append(thresholds, threshold)
		}

		kpis = append(kpis, KPI{
			Name:        yamlKPI.Name,
			Description: yamlKPI.Description,
			PromQL:      yamlKPI.Promql,
			Thresholds:  thresholds,
		})
	}

	return kpis, nil
}

func kpiThresholdFromYaml(threshold YamlThreshold) (Threshold, error) {
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
		return Threshold{}, fmt.Errorf("could not find threshold value: one of [lt, lte, eq, neq, gte, gt] must be provided as a float",)
	}

	if err != nil {
		return Threshold{}, err
	}

	return Threshold{
		Level:    threshold.Level,
		Operator: operator,
		Value:    value,
	}, nil
}
