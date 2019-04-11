package indicator

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func ReadFile(indicatorsFile string, opts ...ReadOpt) (Document, error) {
	fileBytes, err := ioutil.ReadFile(indicatorsFile)
	if err != nil {
		return Document{}, err
	}

	indicatorDocument, err := ReadIndicatorDocument(fileBytes, opts...)
	if err != nil {
		return Document{}, err
	}

	validationErrors := ValidateForRegistry(indicatorDocument)
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

	return indicatorDocument, nil
}

func ReadPatchFile(patchFile string) (Patch, error) {
	fileBytes, err := ioutil.ReadFile(patchFile)
	if err != nil {
		return Patch{}, err
	}

	patch, err := ReadPatchBytes(fileBytes)
	if err != nil {
		return Patch{}, err
	}
	return patch, nil
}
