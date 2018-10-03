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
  for _, yamlIndicator := range d.Indicators {
    var thresholds []Threshold
    for _, yamlThreshold := range yamlIndicator.Thresholds {
      threshold, err := thresholdFromYAML(yamlThreshold)
      if err != nil {
        return Document{}, fmt.Errorf("could not convert yaml indicators to indicators: %s", err)
      }

      thresholds = append(thresholds, threshold)
    }

    indicators = append(indicators, Indicator{
      Name:          yamlIndicator.Name,
      PromQL:        yamlIndicator.Promql,
      Thresholds:    thresholds,
      Documentation: yamlIndicator.Documentation,
    })
  }

  var sections []Section
  for idx, s := range d.YAMLDoc.Sections {

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
    })
  }

  documentation := Documentation{
    Title:       d.YAMLDoc.Title,
    Description: d.YAMLDoc.Description,
    Sections:    sections,
    Owner:       d.YAMLDoc.Owner,
  }

  return Document{
    APIVersion:    d.APIVersion,
    Product:       d.Product,
    Version:       d.Version,
    Metadata:      d.Metadata,
    Indicators:    indicators,
    Documentation: documentation,
  }, nil
}

type yamlDocument struct {
  APIVersion string            `yaml:"apiVersion"`
  Product    string            `yaml:"product"`
  Version    string            `yaml:"version"`
  Metadata   map[string]string `yaml:"metadata"`
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
  Title         string   `yaml:"title"`
  Description   string   `yaml:"description"`
  IndicatorRefs []string `yaml:"indicators"`
}

type yamlIndicator struct {
  Name          string            `yaml:"name"`
  Promql        string            `yaml:"promql"`
  Thresholds    []yamlThreshold   `yaml:"thresholds"`
  SLO           string            `yaml:"slo"`
  Documentation map[string]string `yaml:"documentation"`
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

func findIndicator(name string, indicators []Indicator) (Indicator, bool) {
  for _, i := range indicators {
    if i.Name == name {
      return i, true
    }
  }

  return Indicator{}, false
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
  }, nil
}
