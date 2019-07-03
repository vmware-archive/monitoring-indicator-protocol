package indicator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
)

func ReadFile(indicatorsFile string, opts ...ReadOpt) (Document, error) {
	fileBytes, err := ioutil.ReadFile(indicatorsFile)
	if err != nil {
		return Document{}, err
	}

	reader := ioutil.NopCloser(bytes.NewReader(fileBytes))
	doc, err := DocumentFromYAML(reader)
	if err != nil {
		return Document{}, err
	}

	readOptions := getReadOpts(opts)
	doc.OverrideMetadata(readOptions.overrides)
	if readOptions.interpolate {
		doc.Interpolate()
	}

	validationErrors := doc.Validate("v0", "v1alpha1")
	if len(validationErrors) > 0 {
		var errorS strings.Builder
		errorS.WriteString("validation for indicator document failed:\n")
		for _, e := range validationErrors {
			_, err = fmt.Fprintf(&errorS, "- %v\n", e)
			if err != nil {
				errorS.WriteString("failed to parse error\n")
			}
		}
		return Document{}, fmt.Errorf(errorS.String())
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
		return Patch{}, err
	}

	reader := ioutil.NopCloser(bytes.NewReader(fileBytes))
	patch, err := PatchFromYAML(reader)
	if err != nil {
		return Patch{}, err
	}
	return patch, nil
}
