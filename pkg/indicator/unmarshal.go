package indicator

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

func (t *Threshold) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var threshold struct {
		Level string   `yaml:"level"`
		LT    *float64 `yaml:"lt"`
		LTE   *float64 `yaml:"lte"`
		EQ    *float64 `yaml:"eq"`
		NEQ   *float64 `yaml:"neq"`
		GTE   *float64 `yaml:"gte"`
		GT    *float64 `yaml:"gt"`
	}

	err := unmarshal(&threshold)
	if err != nil {
		return err
	}
	t.Level = threshold.Level

	switch {
	case threshold.LT != nil:
		t.Operator = LessThan
		t.Value = *threshold.LT
	case threshold.LTE != nil:
		t.Operator = LessThanOrEqualTo
		t.Value = *threshold.LTE
	case threshold.EQ != nil:
		t.Operator = EqualTo
		t.Value = *threshold.EQ
	case threshold.NEQ != nil:
		t.Operator = NotEqualTo
		t.Value = *threshold.NEQ
	case threshold.GTE != nil:
		t.Operator = GreaterThanOrEqualTo
		t.Value = *threshold.GTE
	case threshold.GT != nil:
		t.Operator = GreaterThan
		t.Value = *threshold.GT
	default:
		t.Operator = Undefined
	}
	return nil
}

type ReadOpt func(options *readOptions)

func DocumentFromYAML(reader io.ReadCloser) (Document, error) {
	var doc Document
	err := yaml.NewDecoder(reader).Decode(&doc)
	if err != nil {
		return Document{}, fmt.Errorf("could not unmarshal indicators: %s", err)
	}
	_ = reader.Close()

	populateDefaultAlert(&doc)
	populateDefaultLayout(&doc)
	populateDefaultPresentation(&doc)

	return doc, nil
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

	errs := ValidateForRegistry(doc)
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
