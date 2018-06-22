package indicator

import (
	"io/ioutil"
	"log"
	"fmt"
)

func ReadFile(indicatorsFile string) (Document, error) {
	fileBytes, err := ioutil.ReadFile(indicatorsFile)
	if err != nil {
		return Document{}, err
	}

	indicatorDocument, err := ReadIndicatorDocument(fileBytes)
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