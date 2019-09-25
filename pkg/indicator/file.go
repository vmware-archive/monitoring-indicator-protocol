package indicator

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	. "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

// Reads the IndicatorDocument in the file with the given name,
// Returns an error if the file can't be read, or the file isn't valid
// YAML parsable as a document, or the document can't be validated.
func ReadFile(indicatorsFile string, opts ...ReadOpt) (IndicatorDocument, error) {
	fileBytes, err := ioutil.ReadFile(indicatorsFile)

	reg := regexp.MustCompile("<%=.*%>")
	fileBytes = reg.ReplaceAll(fileBytes, []byte("<%= ERB REMOVED FOR YAML SAFETY %>"))

	if err != nil {
		return IndicatorDocument{}, err
	}

	reader := ioutil.NopCloser(bytes.NewReader(fileBytes))
	doc, errs := DocumentFromYAML(reader, opts...)

	if len(errs) > 0 {
		var errorString strings.Builder
		errorString.WriteString("validation for indicator document failed:\n")
		for _, e := range errs {
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
