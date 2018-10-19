package indicator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type ReadOpt func(options *readOptions)

func SkipMetadataInterpolation(options *readOptions) {
	options.interpolate = false
}

func OverrideMetadata(overrideMetadata map[string]string) func(options *readOptions) {
	return func(options *readOptions) {
		for k, v := range overrideMetadata {
			options.overrides[k] = v
		}
	}
}

func ReadIndicatorDocument(yamlBytes []byte, opts ...ReadOpt) (Document, error) {
	readOptions := getReadOpts(opts)

	if readOptions.interpolate {
		metadata, err := readMetadata(yamlBytes)
		if err != nil {
			return Document{}, fmt.Errorf("could not read metadata: %s", err)
		}

		yamlBytes = fillInMetadata(metadata, readOptions.overrides, yamlBytes)
	}

	var d yamlDocument

	err := yaml.Unmarshal(yamlBytes, &d)
	if err != nil {
		return Document{}, fmt.Errorf("could not unmarshal indicators: %s", err)
	}

	for k, v := range readOptions.overrides {
		d.Metadata[k] = v
	}

	var indicators []Indicator
	for idx, yamlIndicator := range d.Indicators {
		var thresholds []Threshold
		for _, yamlThreshold := range yamlIndicator.Thresholds {
			threshold, err := thresholdFromYAML(yamlThreshold)
			if err != nil {
				return Document{}, fmt.Errorf("could not convert yaml indicators to indicators: %s", err)
			}

			thresholds = append(thresholds, threshold)
		}

		var slo float64
		if yamlIndicator.SLO != "" {
			slo, err = strconv.ParseFloat(yamlIndicator.SLO, 64)
			if err != nil {
				return Document{}, fmt.Errorf("indicators[%d] slo unparsable: %s\n", idx, err)
			}
		}

		indicators = append(indicators, Indicator{
			Name:          yamlIndicator.Name,
			PromQL:        yamlIndicator.Promql,
			Thresholds:    thresholds,
			SLO:           slo,
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

func ParseMetadata(input string) map[string]string {
	metadata := map[string]string{}

	for _, pair := range strings.Split(input, ",") {
		v := strings.Split(pair, "=")
		if len(v) > 1 {
			metadata[v[0]] = v[1]
		}
	}

	return metadata
}

func getReadOpts(optionsFuncs []ReadOpt) readOptions {
	options := readOptions{
		interpolate: true,
		overrides:   map[string]string{},
	}

	for _, fn := range optionsFuncs {
		fn(&options)
	}

	return options
}

type readOptions struct {
	interpolate bool
	overrides   map[string]string
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

func readMetadata(document []byte) (map[string]string, error) {
	var d struct {
		Metadata map[string]string `yaml:"metadata"`
	}

	err := yaml.Unmarshal(document, &d)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal metadata: %s", err)
	}

	return d.Metadata, nil
}

func fillInMetadata(documentMetadata map[string]string, overrideMetadata map[string]string, documentBytes []byte) []byte {

	for k, v := range overrideMetadata {
		documentMetadata[k] = v
	}

	for k, v := range documentMetadata {
		documentBytes = bytes.Replace(documentBytes, []byte("$"+k), []byte(v), -1)
	}

	return documentBytes
}
