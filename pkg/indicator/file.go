package indicator

import (
	"fmt"
	"io/ioutil"
	"log"
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

	validationErrors := Validate(indicatorDocument)
	if len(validationErrors) > 0 {

		for _, e := range validationErrors {
			log.Printf("- %s \n", e.Error())
		}

		return Document{}, fmt.Errorf("validation for indicator file failed - [%+v]", validationErrors)
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
