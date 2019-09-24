package indicator

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	. "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

// Reads the IndicatorDocument in the file with the given name,
// Returns an error if the file can't be read, or the file isn't valid
// YAML parsable as a document, or the document can't be validated.
func ReadFile(indicatorsFile string, opts ...ReadOpt) (IndicatorDocument, error) {
	fileBytes, err := ioutil.ReadFile(indicatorsFile)
	if err != nil {
		return IndicatorDocument{}, err
	}

	reader := ioutil.NopCloser(bytes.NewReader(fileBytes))
	doc, err := DocumentFromYAML(reader, opts...)
	if err != nil {
		return IndicatorDocument{}, err
	}

	readOptions := getReadOpts(opts)
	doc.OverrideMetadata(readOptions.overrides)
	if readOptions.interpolate {
		doc.Interpolate()
	}

	validationErrors := doc.Validate(api_versions.V0, api_versions.V1)
	if len(validationErrors) > 0 {
		var errorString strings.Builder
		errorString.WriteString("validation for indicator document failed:\n")
		for _, e := range validationErrors {
			_, err = fmt.Fprintf(&errorString, "- %v\n", e)
			if err != nil {
				errorString.WriteString("failed to parse error\n")
			}
		}
		return IndicatorDocument{}, errors.New(errorString.String())
	}

	return doc, nil
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

func ReadPatchFile(patchFile string) (Patch, error) {
	fileBytes, err := ioutil.ReadFile(patchFile)
	if err != nil {
		return Patch{}, errors.New("could not read patch file")
	}

	reader := ioutil.NopCloser(bytes.NewReader(fileBytes))
	patch, err := PatchFromYAML(reader)
	if err != nil {
		return Patch{}, errors.New("could not unmarshal patch file")
	}
	return patch, nil
}
