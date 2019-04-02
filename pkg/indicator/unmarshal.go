package indicator

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cppforlife/go-patch/patch"
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

func ProcessDocument(patches []Patch, documentBytes []byte) (Document, []error) {
	patchedDocBytes, err := ApplyPatches(patches, documentBytes)
	if err != nil {
		return Document{}, []error{err}
	}

	doc, err := ReadIndicatorDocument(patchedDocBytes)
	if err != nil {
		return Document{}, []error{err}
	}

	errs := Validate(doc)
	if len(errs) > 0 {
		return Document{}, errs
	}

	return doc, nil
}

func ApplyPatches(patches []Patch, documentBytes []byte) ([]byte, error) {
	_, err := readMetadata(documentBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("could not read document metadata: %s", err)
	}
	var document interface{}
	err = yaml.Unmarshal(documentBytes, &document)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to unmarshal document for patching: %s", err)
	}

	for _, p := range patches {
		if MatchDocument(p.Match, documentBytes) {
			ops, err := patch.NewOpsFromDefinitions(p.Operations)
			if err != nil {
				log.Print(fmt.Errorf("failed to parse patch operations: %s", err))
				continue
			}
			for i, o := range ops {
				var tempDocument interface{}
				tempDocument, err = o.Apply(document)
				if err != nil {
					od := p.Operations[i]
					log.Print(fmt.Errorf("failed to apply patch operation %s %s: %s", od.Type, *od.Path, err))
					continue
				}
				document = tempDocument
			}
		}
	}

	patched, err := yaml.Marshal(document)
	if err != nil {
		return []byte{}, err
	}
	return patched, nil
}

func MatchDocument(criteria Match, documentBytes []byte) bool {
	product, err := readProductInfo(documentBytes)
	if err != nil {
		return false
	}

	if criteria.Name != nil && *criteria.Name != product.Name {
		return false
	}
	if criteria.Version != nil && *criteria.Version != product.Version {
		return false
	}

	if criteria.Metadata != nil {
		metadata, err := readMetadata(documentBytes)
		if err != nil {
			return false
		}

		if !reflect.DeepEqual(metadata, criteria.Metadata) {
			return false
		}
	}

	return true
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
	for i, yamlIndicator := range d.Indicators {
		var thresholds []Threshold
		for _, yamlThreshold := range yamlIndicator.Thresholds {
			threshold, err := thresholdFromYAML(yamlThreshold)
			if err != nil {
				return Document{}, fmt.Errorf("could not convert yaml indicator[%v]: %s", i, err)
			}

			thresholds = append(thresholds, threshold)
		}

		p, err := presentationFromYAML(yamlIndicator.Presentation)
		if err != nil {
			return Document{}, fmt.Errorf("could not convert yaml indicator[%v]: %s", i, err)
		}

		indicators = append(indicators, Indicator{
			Name:          yamlIndicator.Name,
			PromQL:        yamlIndicator.Promql,
			Thresholds:    thresholds,
			Alert:         alertFromYAML(yamlIndicator.Alert),
			Presentation:  p,
			Documentation: yamlIndicator.Documentation,
		})
	}

	layout, err := getLayout(d.YAMLLayout, indicators)
	if err != nil {
		return Document{}, err
	}

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

func ReadPatchBytes(yamlBytes []byte) (Patch, error) {
	p := yamlPatch{}
	err := yaml.Unmarshal(yamlBytes, &p)

	if err != nil {
		return Patch{}, fmt.Errorf("unable to parse bytes: %s\n", err)
	}

	return Patch{
		APIVersion: p.APIVersion,
		Match: Match{
			Name:     p.Match.Product.Name,
			Version:  p.Match.Product.Version,
			Metadata: p.Match.Metadata,
		},
		Operations: p.Operations,
	}, nil
}

func getLayout(l *yamlLayout, indicators []Indicator) (Layout, error) {
	var sections []Section
	if l == nil {
		return Layout{
			Sections: []Section{{
				Title:      "Metrics",
				Indicators: indicators,
			}},
		}, nil
	}

	for idx, s := range l.Sections {
		var sectionIndicators []Indicator
		for iIdx, i := range s.IndicatorRefs {
			indic, ok := findIndicator(i, indicators)
			if !ok {
				return Layout{}, fmt.Errorf("documentation.sections[%d].indicators[%d] references non-existent indicator", idx, iIdx)
			}

			sectionIndicators = append(sectionIndicators, indic)
		}

		sections = append(sections, Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  sectionIndicators,
		})
	}

	return Layout{
		Title:       l.Title,
		Description: l.Description,
		Owner:       l.Owner,
		Sections:    sections,
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
	Product    yamlProduct       `yaml:"product"`
	Metadata   map[string]string `yaml:"metadata"`
	Indicators []yamlIndicator   `yaml:"indicators"`
	YAMLLayout *yamlLayout       `yaml:"layout"`
}

type yamlProduct struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type yamlLayout struct {
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
	Alert         yamlAlert         `yaml:"alert"`
	Documentation map[string]string `yaml:"documentation"`
	Presentation  yamlPresentation  `yaml:"presentation"`
}

type yamlAlert struct {
	For  string
	Step string
}

type yamlThreshold struct {
	Level string `yaml:"level"`
	LT    string `yaml:"lt"`
	LTE   string `yaml:"lte"`
	EQ    string `yaml:"eq"`
	NEQ   string `yaml:"neq"`
	GTE   string `yaml:"gte"`
	GT    string `yaml:"gt"`
}

type yamlPresentation struct {
	ChartType    ChartType     `yaml:"chartType"`
	CurrentValue bool          `yaml:"currentValue"`
	Frequency    time.Duration `yaml:"frequency"`
	Labels       []string      `yaml:"labels"`
	Units        string        `yaml:"units"`
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

func presentationFromYAML(p yamlPresentation) (*Presentation, error) {
	defaultValue := yamlPresentation{}
	if reflect.DeepEqual(p, defaultValue) {
		return &Presentation{
			ChartType:    StepChart,
			CurrentValue: false,
			Frequency:    0,
			Labels:       []string{},
			Units:        "",
		}, nil
	}

	chartType := p.ChartType
	if chartType == "" {
		chartType = StepChart
	}

	return &Presentation{
		ChartType:    chartType,
		CurrentValue: p.CurrentValue,
		Frequency:    p.Frequency,
		Labels:       p.Labels,
		Units:        p.Units,
	}, nil
}

func alertFromYAML(a yamlAlert) Alert {
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

func readProductInfo(documentBytes []byte) (yamlProduct, error) {
	var document struct {
		Product yamlProduct `yaml:"product"`
	}

	err := yaml.Unmarshal(documentBytes, &document)
	if err != nil {
		return yamlProduct{}, fmt.Errorf("could not unmarshal metadata: %s", err)
	}

	return document.Product, nil
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
