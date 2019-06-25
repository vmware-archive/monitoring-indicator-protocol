package indicator

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"reflect"
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

type yamlProduct struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
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

