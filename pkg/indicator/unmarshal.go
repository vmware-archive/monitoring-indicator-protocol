package indicator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"github.com/ghodss/yaml"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

type ReadOpt func(options *readOptions)

func DocumentFromYAML(r io.ReadCloser, opts ...ReadOpt) (v1.IndicatorDocument, []error) {
	docBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return v1.IndicatorDocument{}, []error{err}
	}

	readOptions := getReadOpts(opts)
	if len(readOptions.overrides) > 0 {
		docBytes, err = overrideMetadataBytes(docBytes, readOptions.overrides)
		if err != nil {
			return v1.IndicatorDocument{}, []error{err}
		}
	}
	if readOptions.interpolate {
		docBytes, err = interpolateBytes(docBytes)
		if err != nil {
			return v1.IndicatorDocument{}, []error{err}
		}
	}

	apiVersion, err := ApiVersionFromYAML(docBytes)
	if err != nil {
		return v1.IndicatorDocument{}, []error{err}
	}

	var doc v1.IndicatorDocument
	switch apiVersion {
	case api_versions.V1:
		err = yaml.Unmarshal(docBytes, &doc)
	default:
		err = fmt.Errorf("invalid apiVersion, supported versions are: [indicatorprotocol.io/v1]")
	}

	if err != nil {
		return v1.IndicatorDocument{}, []error{err}
	}

	v1.PopulateDefaults(&doc)

	validationErrors := doc.Validate()
	if len(validationErrors) > 0 {
		return v1.IndicatorDocument{}, validationErrors
	}

	return doc, []error{}
}

// Assuming the given bytes are yaml, upserts the given key/value pairs into the `metadata.labels` of the given
// yaml.
func overrideMetadataBytes(docBytes []byte, overrides map[string]string) ([]byte, error) {
	metadata, err := MetadataFromYAML(ioutil.NopCloser(bytes.NewReader(docBytes)))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse metadata, %s", err)
	}

	for _, label := range sortLabels(overrides) {
		metadata[label] = overrides[label]
	}

	return writeMetadataToYaml(docBytes, metadata)
}

type byLargestLength []string

func (s byLargestLength) Len() int {
	return len(s)
}
func (s byLargestLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLargestLength) Less(i, j int) bool {
	return len(s[i]) > len(s[j])
}

func sortLabels(labels map[string]string) []string {
	sorted := make([]string, 0)
	for k := range labels {
		sorted = append(sorted, k)
	}

	sort.Sort(byLargestLength(sorted))
	return sorted
}

func writeMetadataToYaml(docBytes []byte, metadata map[string]string) ([]byte, error) {
	apiVersion, err := ApiVersionFromYAML(docBytes)
	if err != nil {
		return nil, err
	}
	switch apiVersion {
	case api_versions.V1:
		var docMap map[string]interface{}
		err := yaml.Unmarshal(docBytes, &docMap)
		if err != nil {
			return nil, err
		}

		m, ok := docMap["metadata"]
		if !ok {
			return nil, errors.New("error writing metadata to document, document has no metadata key")
		}
		meta, ok := m.(map[string]interface{})
		if !ok {
			return nil, errors.New("error writing metadata to document, document metadata is not a mapping")
		}
		meta["labels"] = metadata

		return yaml.Marshal(docMap)
	default:
		return nil, errors.New("invalid apiVersion")
	}
}

func interpolateBytes(docBytes []byte) ([]byte, error) {
	metadata, err := MetadataFromYAML(ioutil.NopCloser(bytes.NewReader(docBytes)))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse metadata, %s", err)
	}

	for key, value := range metadata {
		regString := fmt.Sprintf(`(\$%s)(\b|\_|$)|(\$\{%s\})`, key, key)
		regex := regexp.MustCompile(regString)
		docBytes = regex.ReplaceAll(docBytes, []byte(fmt.Sprintf("%s$2", value)))
	}

	return docBytes, nil
}

func ApiVersionFromYAML(docBytes []byte) (string, error) {
	var d struct {
		ApiVersion string `yaml:"apiVersion"`
	}
	err := yaml.Unmarshal(docBytes, &d)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal apiVersion, check that document contains valid YAML")
	}
	return d.ApiVersion, nil
}

func KindFromYAML(fileBytes []byte) (string, error) {
	var f struct{ Kind string }

	err := yaml.Unmarshal(fileBytes, &f)
	if err != nil {
		return "", err
	}

	return f.Kind, nil
}

func PatchFromYAML(reader io.ReadCloser) (Patch, error) {
	var yamlPatch yamlPatch
	patchBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return Patch{}, fmt.Errorf("could not read patch: %s", err)
	}
	err = yaml.Unmarshal(patchBytes, &yamlPatch)
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

func ProductFromYAML(reader io.ReadCloser) (v1.Product, error) {
	docBytes, err := ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return v1.Product{}, fmt.Errorf("could not read document")
	}

	apiVersion, err := ApiVersionFromYAML(docBytes)
	var product v1.Product
	switch apiVersion {
	case api_versions.V1:
		var d struct {
			Spec struct {
				Product v1.Product
			}
		}
		err = yaml.Unmarshal(docBytes, &d)
		product = d.Spec.Product
	}

	if err != nil {
		return v1.Product{}, errors.New("could not unmarshal product information")
	}

	return product, nil
}

func MetadataFromYAML(reader io.ReadCloser) (map[string]string, error) {
	docBytes, err := ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read document")
	}

	apiVersion, err := ApiVersionFromYAML(docBytes)
	if err != nil {
		return nil, fmt.Errorf("could not read apiVersion: %s", err)
	}
	var metadata map[string]string
	switch apiVersion {
	case api_versions.V1:
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

func ProcessDocument(patches []Patch, documentBytes []byte) (v1.IndicatorDocument, []error) {
	patchedDocBytes, err := ApplyPatches(patches, documentBytes)
	if err != nil {
		log.Print("failed to apply patches to document")
		return v1.IndicatorDocument{}, []error{err}
	}

	reader := ioutil.NopCloser(bytes.NewReader(patchedDocBytes))
	doc, errs := DocumentFromYAML(reader)
	if len(errs) > 0 {
		log.Print("failed to unmarshal document")
		return v1.IndicatorDocument{}, errs
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
