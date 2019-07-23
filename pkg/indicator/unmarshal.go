package indicator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"github.com/ghodss/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

type ReadOpt func(options *readOptions)

func DocumentFromYAML(r io.ReadCloser) (v1alpha1.IndicatorDocument, error) {
	docBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return v1alpha1.IndicatorDocument{}, err
	}

	apiVersion, err := apiVersionFromYAML(docBytes)
	if err != nil {
		return v1alpha1.IndicatorDocument{}, err
	}

	var doc v1alpha1.IndicatorDocument
	switch apiVersion {
	case "v0":
		log.Print("WARNING: apiVersion v0 will be deprecated in future releases")
		doc, err = v0documentFromBytes(docBytes)
	case "apps.pivotal.io/v1alpha1":
		err = yaml.Unmarshal(docBytes, &doc)
	default:
		err = fmt.Errorf("invalid apiVersion, supported versions are: v0, apps.pivotal.io/v1alpha1")
	}

	if err != nil {
		return v1alpha1.IndicatorDocument{}, err
	}

	populateDefaultAlert(&doc)
	populateDefaultLayout(&doc)
	populateDefaultPresentation(&doc)

	return doc, nil
}

func v0documentFromBytes(yamlBytes []byte) (v1alpha1.IndicatorDocument, error) {
	var d v0yamlDocument

	err := yaml.Unmarshal(yamlBytes, &d)
	if err != nil {
		return v1alpha1.IndicatorDocument{}, fmt.Errorf("could not unmarshal indicator document")
	}

	var indicators []v1alpha1.IndicatorSpec
	for indicatorIndex, yamlIndicator := range d.Indicators {
		var thresholds []v1alpha1.Threshold
		for thresholdIndex, yamlThreshold := range yamlIndicator.Thresholds {
			threshold, err := v0thresholdFromYAML(yamlThreshold)
			if err != nil {
				return v1alpha1.IndicatorDocument{}, fmt.Errorf("could not unmarshal threshold %v in indicator %v", thresholdIndex, indicatorIndex)
			}

			thresholds = append(thresholds, threshold)
		}

		p := v0presentationFromYAML(yamlIndicator.Presentation)

		indicators = append(indicators, v1alpha1.IndicatorSpec{
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

	return v1alpha1.IndicatorDocument{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v0",
			Kind:       "IndicatorDocument",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: d.Metadata,
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{
				Name:    d.Product.Name,
				Version: d.Product.Version,
			},
			Indicators: indicators,
			Layout:     layout,
		},
	}, nil
}

func getLayout(l *v0yamlLayout) v1alpha1.Layout {
	if l == nil {
		return v1alpha1.Layout{}
	}
	sections := make([]v1alpha1.Section, 0)

	for _, s := range l.Sections {
		sections = append(sections, v1alpha1.Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  s.IndicatorRefs,
		})
	}

	return v1alpha1.Layout{
		Title:       l.Title,
		Description: l.Description,
		Owner:       l.Owner,
		Sections:    sections,
	}
}

func v0thresholdFromYAML(threshold v0yamlThreshold) (v1alpha1.Threshold, error) {
	var operator v1alpha1.ThresholdOperator
	var value float64
	var err error

	switch {
	case threshold.LT != "":
		operator = v1alpha1.LessThan
		value, err = strconv.ParseFloat(threshold.LT, 64)
	case threshold.LTE != "":
		operator = v1alpha1.LessThanOrEqualTo
		value, err = strconv.ParseFloat(threshold.LTE, 64)
	case threshold.EQ != "":
		operator = v1alpha1.EqualTo
		value, err = strconv.ParseFloat(threshold.EQ, 64)
	case threshold.NEQ != "":
		operator = v1alpha1.NotEqualTo
		value, err = strconv.ParseFloat(threshold.NEQ, 64)
	case threshold.GTE != "":
		operator = v1alpha1.GreaterThanOrEqualTo
		value, err = strconv.ParseFloat(threshold.GTE, 64)
	case threshold.GT != "":
		operator = v1alpha1.GreaterThan
		value, err = strconv.ParseFloat(threshold.GT, 64)
	default:
		operator = v1alpha1.Undefined
	}

	if err != nil {
		return v1alpha1.Threshold{}, err
	}

	return v1alpha1.Threshold{
		Level:    threshold.Level,
		Operator: operator,
		Value:    value,
	}, nil
}

func v0presentationFromYAML(p *v0yamlPresentation) v1alpha1.Presentation {
	if p == nil {
		return v1alpha1.Presentation{
			ChartType:    v1alpha1.StepChart,
			CurrentValue: false,
			Frequency:    0,
			Labels:       []string{},
			Units:        "",
		}
	}

	chartType := p.ChartType
	if chartType == "" {
		chartType = v1alpha1.StepChart
	}

	return v1alpha1.Presentation{
		ChartType:    chartType,
		CurrentValue: p.CurrentValue,
		Frequency:    p.Frequency,
		Labels:       p.Labels,
		Units:        p.Units,
	}
}

func v0alertFromYAML(a v0yamlAlert) v1alpha1.Alert {
	alertFor, alertStep := a.For, a.Step
	if alertFor == "" {
		alertFor = "1m"
	}
	if alertStep == "" {
		alertStep = "1m"
	}

	return v1alpha1.Alert{
		For:  alertFor,
		Step: alertStep,
	}
}

func v0serviceLevelFromYAML(level *v0yamlServiceLevel) *v1alpha1.ServiceLevel {
	if level == nil {
		return nil
	}
	return &v1alpha1.ServiceLevel{
		Objective: level.Objective,
	}
}

type v0yamlDocument struct {
	APIVersion string            `json:"apiVersion"`
	Product    v0yamlProduct     `json:"product"`
	Metadata   map[string]string `json:"metadata"`
	Indicators []v0yamlIndicator `json:"indicators"`
	YAMLLayout *v0yamlLayout     `json:"layout"`
}

type v0yamlProduct struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type v0yamlLayout struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Sections    []v0yamlSection `json:"sections"`
	Owner       string          `json:"owner"`
}

type v0yamlSection struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	IndicatorRefs []string `json:"indicators"`
}

type v0yamlIndicator struct {
	Name          string              `json:"name"`
	Promql        string              `json:"promql"`
	Thresholds    []v0yamlThreshold   `json:"thresholds"`
	Alert         v0yamlAlert         `json:"alert"`
	ServiceLevel  *v0yamlServiceLevel `json:"serviceLevel"`
	Documentation map[string]string   `json:"documentation"`
	Presentation  *v0yamlPresentation `json:"presentation"`
}

type v0yamlServiceLevel struct {
	Objective float64 `json:"objective"`
}

type v0yamlAlert struct {
	For  string
	Step string
}

type v0yamlThreshold struct {
	Level string `json:"level"`
	LT    string `json:"lt"`
	LTE   string `json:"lte"`
	EQ    string `json:"eq"`
	NEQ   string `json:"neq"`
	GTE   string `json:"gte"`
	GT    string `json:"gt"`
}

type v0yamlPresentation struct {
	ChartType    v1alpha1.ChartType `json:"chartType"`
	CurrentValue bool      `json:"currentValue"`
	Frequency    int64     `json:"frequency"`
	Labels       []string  `json:"labels"`
	Units        string    `json:"units"`
}

func apiVersionFromYAML(docBytes []byte) (string, error) {
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
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return Patch{}, fmt.Errorf("could not read patch: %s", err)
	}
	err = yaml.Unmarshal(bytes, &yamlPatch)
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

func ProductFromYAML(reader io.ReadCloser) (v1alpha1.Product, error) {
	docBytes, err := ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return v1alpha1.Product{}, fmt.Errorf("could not read document")
	}

	apiVersion, err := apiVersionFromYAML(docBytes)
	var product v1alpha1.Product
	switch apiVersion {
	case "v0":
		var d struct {
			Product v1alpha1.Product
		}
		err = yaml.Unmarshal(docBytes, &d)
		product = d.Product
	case "apps.pivotal.io/v1alpha1":
		var d struct {
			Spec struct {
				Product v1alpha1.Product
			}
		}
		err = yaml.Unmarshal(docBytes, &d)
		product = d.Spec.Product
	}

	if err != nil {
		return v1alpha1.Product{}, errors.New("could not unmarshal product information")
	}

	return product, nil
}

func MetadataFromYAML(reader io.ReadCloser) (map[string]string, error) {
	docBytes, err := ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read document")
	}

	apiVersion, err := apiVersionFromYAML(docBytes)
	var metadata map[string]string
	switch apiVersion {
	case "v0":
		var d struct {
			Metadata map[string]string
		}
		err = yaml.Unmarshal(docBytes, &d)
		metadata = d.Metadata
	case "apps.pivotal.io/v1alpha1":
		var d struct {
			Metadata struct {
				Labels map[string]string
			}
		}
		err = yaml.Unmarshal(docBytes, &d)
		metadata = d.Metadata.Labels
	}

	if err != nil {
		return map[string]string{}, fmt.Errorf("could not unmarshal metadata")
	}
	_ = reader.Close()

	return metadata, nil
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

func populateDefaultPresentation(doc *v1alpha1.IndicatorDocument) {
	for i, indicator := range doc.Spec.Indicators {
		if indicator.Presentation.ChartType == "" {
			doc.Spec.Indicators[i].Presentation.ChartType = "step"
		}
		if indicator.Presentation.Labels == nil {
			doc.Spec.Indicators[i].Presentation.Labels = []string{}
		}
	}
}

func populateDefaultLayout(doc *v1alpha1.IndicatorDocument) {
	if doc.Spec.Layout.Sections == nil {
		indicatorNames := make([]string, 0)
		for _, indicator := range doc.Spec.Indicators {
			indicatorNames = append(indicatorNames, indicator.Name)
		}
		doc.Spec.Layout.Sections = []v1alpha1.Section{
			{
				Title:      "Metrics",
				Indicators: indicatorNames,
			},
		}
	}
}

func populateDefaultAlert(doc *v1alpha1.IndicatorDocument) {
	for i, indicator := range doc.Spec.Indicators {
		if indicator.Alert.For == "" {
			doc.Spec.Indicators[i].Alert.For = "1m"
		}
		if indicator.Alert.Step == "" {
			doc.Spec.Indicators[i].Alert.Step = "1m"
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

func ProcessDocument(patches []Patch, documentBytes []byte) (v1alpha1.IndicatorDocument, []error) {
	patchedDocBytes, err := ApplyPatches(patches, documentBytes)
	if err != nil {
		log.Print("failed to apply patches to document")
		return v1alpha1.IndicatorDocument{}, []error{err}
	}

	reader2 := ioutil.NopCloser(bytes.NewReader(patchedDocBytes))
	doc, err := DocumentFromYAML(reader2)
	if err != nil {
		log.Print("failed to unmarshal document")
		return v1alpha1.IndicatorDocument{}, []error{err}
	}
	doc.Interpolate()

	errs := doc.Validate("v0", "apps.pivotal.io/v1alpha1")
	if len(errs) > 0 {
		log.Print("document validation failed")
		for _, e := range errs {
			log.Printf("- %s \n", e.Error())
		}
		return v1alpha1.IndicatorDocument{}, errs
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
