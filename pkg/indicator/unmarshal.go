package indicator

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

func (t *Threshold) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var threshold struct {
		Level    string  `yaml:"level"`
		Operator string  `yaml:"operator"`
		Value    float64 `yaml:"value"`
	}

	err := unmarshal(&threshold)
	if err != nil {
		return err
	}
	t.Level = threshold.Level
	t.Value = threshold.Value

	switch {
	case threshold.Operator == "lt":
		t.Operator = LessThan
	case threshold.Operator == "lte":
		t.Operator = LessThanOrEqualTo
	case threshold.Operator == "eq":
		t.Operator = EqualTo
	case threshold.Operator == "neq":
		t.Operator = NotEqualTo
	case threshold.Operator == "gte":
		t.Operator = GreaterThanOrEqualTo
	case threshold.Operator == "gt":
		t.Operator = GreaterThan
	default:
		t.Operator = Undefined
	}
	return nil
}

type ReadOpt func(options *readOptions)

func DocumentFromYAML(reader io.ReadCloser) (Document, error) {
	var doc Document

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return Document{}, fmt.Errorf("could not read indicator document")
	}
	docBytes := buf.Bytes()
	_ = reader.Close()

	apiVersion, err := v0apiVersionFromYAML(docBytes)
	if err != nil {
		return Document{}, err
	}
	if apiVersion == "v0" {
		doc, err = v0documentFromBytes(buf.Bytes())
	} else if apiVersion == "v1alpha1" {
		err = yaml.Unmarshal(docBytes, &doc)
	} else {
		err = fmt.Errorf("invalid apiVersion, supported versions are: v0, v1alpha1")
	}
	if err != nil {
		return Document{}, err
	}

	populateDefaultAlert(&doc)
	populateDefaultLayout(&doc)
	populateDefaultPresentation(&doc)

	return doc, nil
}

func v0documentFromBytes(yamlBytes []byte) (Document, error) {
	var d v0yamlDocument

	err := yaml.Unmarshal(yamlBytes, &d)
	if err != nil {
		return Document{}, fmt.Errorf("could not unmarshal indicator document")
	}

	var indicators []Indicator
	for indicatorIndex, yamlIndicator := range d.Indicators {
		var thresholds []Threshold
		for thresholdIndex, yamlThreshold := range yamlIndicator.Thresholds {
			threshold, err := v0thresholdFromYAML(yamlThreshold)
			if err != nil {
				return Document{}, fmt.Errorf("could not unmarshal threshold %v in indicator %v", thresholdIndex, indicatorIndex)
			}

			thresholds = append(thresholds, threshold)
		}

		p := v0presentationFromYAML(yamlIndicator.Presentation)

		indicators = append(indicators, Indicator{
			Name:          yamlIndicator.Name,
			PromQL:        yamlIndicator.Promql,
			Thresholds:    thresholds,
			Alert:         v0alertFromYAML(yamlIndicator.Alert),
			ServiceLevel:  v0serviceLevelFromYAML(yamlIndicator.ServiceLevel),
			Presentation:  p,
			Documentation: yamlIndicator.Documentation,
		})
	}

	layout := getLayout(d.YAMLLayout)

	return Document{
		APIVersion: d.APIVersion,
		Product: Product{
			Name:    d.Product.Name,
			Version: d.Product.Version,
		},
		Metadata:   d.Metadata,
		Indicators: indicators,
		Layout:     layout,
	}, nil
}

func getLayout(l *v0yamlLayout) Layout {
	var sections []Section
	if l == nil {
		return Layout{}
	}

	for _, s := range l.Sections {
		sections = append(sections, Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  s.IndicatorRefs,
		})
	}

	return Layout{
		Title:       l.Title,
		Description: l.Description,
		Owner:       l.Owner,
		Sections:    sections,
	}
}

func v0thresholdFromYAML(threshold v0yamlThreshold) (Threshold, error) {
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
		operator = Undefined
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

func v0presentationFromYAML(p *v0yamlPresentation) Presentation {
	if p == nil {
		return Presentation{
			ChartType:    StepChart,
			CurrentValue: false,
			Frequency:    0,
			Labels:       []string{},
			Units:        "",
		}
	}

	chartType := p.ChartType
	if chartType == "" {
		chartType = StepChart
	}

	return Presentation{
		ChartType:    chartType,
		CurrentValue: p.CurrentValue,
		Frequency:    p.Frequency,
		Labels:       p.Labels,
		Units:        p.Units,
	}
}

func v0alertFromYAML(a v0yamlAlert) Alert {
	alertFor, alertStep := a.For, a.Step
	if alertFor == "" {
		alertFor = "1m"
	}
	if alertStep == "" {
		alertStep = "1m"
	}

	return Alert{
		For:  alertFor,
		Step: alertStep,
	}
}

func v0serviceLevelFromYAML(level *v0yamlServiceLevel) *ServiceLevel {
	if level == nil {
		return nil
	}
	return &ServiceLevel{
		Objective: level.Objective,
	}
}

type v0yamlDocument struct {
	APIVersion string            `yaml:"apiVersion"`
	Product    v0yamlProduct     `yaml:"product"`
	Metadata   map[string]string `yaml:"metadata"`
	Indicators []v0yamlIndicator `yaml:"indicators"`
	YAMLLayout *v0yamlLayout     `yaml:"layout"`
}

type v0yamlProduct struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type v0yamlLayout struct {
	Title       string          `yaml:"title"`
	Description string          `yaml:"description"`
	Sections    []v0yamlSection `yaml:"sections"`
	Owner       string          `yaml:"owner"`
}

type v0yamlSection struct {
	Title         string   `yaml:"title"`
	Description   string   `yaml:"description"`
	IndicatorRefs []string `yaml:"indicators"`
}

type v0yamlIndicator struct {
	Name          string              `yaml:"name"`
	Promql        string              `yaml:"promql"`
	Thresholds    []v0yamlThreshold   `yaml:"thresholds"`
	Alert         v0yamlAlert         `yaml:"alert"`
	ServiceLevel  *v0yamlServiceLevel `yaml:"serviceLevel"`
	Documentation map[string]string   `yaml:"documentation"`
	Presentation  *v0yamlPresentation `yaml:"presentation"`
}

type v0yamlServiceLevel struct {
	Objective float64 `yaml:"objective"`
}

type v0yamlAlert struct {
	For  string
	Step string
}

type v0yamlThreshold struct {
	Level string `yaml:"level"`
	LT    string `yaml:"lt"`
	LTE   string `yaml:"lte"`
	EQ    string `yaml:"eq"`
	NEQ   string `yaml:"neq"`
	GTE   string `yaml:"gte"`
	GT    string `yaml:"gt"`
}

type v0yamlPresentation struct {
	ChartType    ChartType `yaml:"chartType"`
	CurrentValue bool      `yaml:"currentValue"`
	Frequency    int64     `yaml:"frequency"`
	Labels       []string  `yaml:"labels"`
	Units        string    `yaml:"units"`
}

func v0apiVersionFromYAML(docBytes []byte) (string, error) {
	var d struct {
		ApiVersion string `yaml:"apiVersion"`
	}
	err := yaml.Unmarshal(docBytes, &d)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal apiVersion")
	}
	return d.ApiVersion, nil
}

func PatchFromYAML(reader io.ReadCloser) (Patch, error) {
	var yamlPatch yamlPatch
	err := yaml.NewDecoder(reader).Decode(&yamlPatch)
	if err != nil {
		return Patch{}, fmt.Errorf("could not unmarshal patch: %s", err)
	}
	_ = reader.Close()

	return Patch{
		APIVersion: yamlPatch.APIVersion,
		Match: Match{
			Name:     yamlPatch.Match.Product.Name,
			Version:  yamlPatch.Match.Product.Version,
			Metadata: yamlPatch.Match.Metadata,
		},
		Operations: yamlPatch.Operations,
	}, nil
}

func ProductFromYAML(reader io.ReadCloser) (Product, error) {
	var d struct {
		Product Product `yaml:"product"`
	}
	err := yaml.NewDecoder(reader).Decode(&d)
	if err != nil {
		return Product{}, fmt.Errorf("could not unmarshal product information: %s", err)
	}
	_ = reader.Close()

	return d.Product, nil
}

func MetadataFromYAML(reader io.ReadCloser) (map[string]string, error) {
	var d struct {
		Metadata map[string]string `yaml:"metadata"`
	}
	err := yaml.NewDecoder(reader).Decode(&d)
	if err != nil {
		return map[string]string{}, fmt.Errorf("could not unmarshal metadata: %s", err)
	}
	_ = reader.Close()

	return d.Metadata, nil
}

type yamlPatch struct {
	APIVersion string               `yaml:"apiVersion"`
	Match      yamlMatch            `yaml:"match"`
	Operations []patch.OpDefinition `yaml:"operations"`
}

type yamlMatch struct {
	Product struct {
		Name    *string `yaml:"name,omitempty"`
		Version *string `yaml:"version,omitempty"`
	} `yaml:"product,omitempty"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

func populateDefaultPresentation(doc *Document) {
	for i, indicator := range doc.Indicators {
		if indicator.Presentation.ChartType == "" {
			doc.Indicators[i].Presentation.ChartType = "step"
		}
		if indicator.Presentation.Labels == nil {
			doc.Indicators[i].Presentation.Labels = []string{}
		}
	}
}

func populateDefaultLayout(doc *Document) {
	if doc.Layout.Sections == nil {
		//noinspection GoPreferNilSlice
		indicatorNames := []string{}
		for _, indicator := range doc.Indicators {
			indicatorNames = append(indicatorNames, indicator.Name)
		}
		doc.Layout.Sections = []Section{
			{
				Title:      "Metrics",
				Indicators: indicatorNames,
			},
		}
	}
}

func populateDefaultAlert(doc *Document) {
	for i, indicator := range doc.Indicators {
		if indicator.Alert.For == "" {
			doc.Indicators[i].Alert.For = "1m"
		}
		if indicator.Alert.Step == "" {
			doc.Indicators[i].Alert.Step = "1m"
		}
	}
}

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

func ProcessDocument(patches []Patch, documentBytes []byte) (Document, []error) {
	patchedDocBytes, err := ApplyPatches(patches, documentBytes)
	if err != nil {
		return Document{}, []error{err}
	}

	reader := ioutil.NopCloser(bytes.NewReader(patchedDocBytes))
	doc, err := DocumentFromYAML(reader)
	if err != nil {
		return Document{}, []error{err}
	}
	doc.Interpolate()

	errs := doc.Validate("v0", "v1alpha1")
	if len(errs) > 0 {
		return Document{}, errs
	}

	return doc, nil
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

type readOptions struct {
	interpolate bool
	overrides   map[string]string
}
