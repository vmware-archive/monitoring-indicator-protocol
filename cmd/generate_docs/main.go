package main

import (
	"os"
	"io/ioutil"
	"log"
	"fmt"

	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"code.cloudfoundry.org/cf-indicators/pkg/docs"
)

func main() {
	fileBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("cannot read file %s: %s", os.Args[1], err)
	}

	indicatorDocument, err := indicator.ReadIndicatorDocument(fileBytes)
	if err != nil {
		log.Fatalf("cannot parse file: %s", err)
	}

	validationErrors := indicator.Validate(indicatorDocument)
	if len(validationErrors) > 0 {

		log.Println("validation for indicator file failed")
		for _, e := range validationErrors {
			log.Printf("- %s \n", e.Error())
		}

		os.Exit(1)
	}

	documentation, err := docs.ConvertIndicatorDocument(indicatorDocument)
	if err != nil {
		log.Fatalf("cannot convert indicator document: %s", err)
	}

	html, err := docs.DocumentToHTML(documentation)
	if err != nil {
		log.Fatalf("cannot render document: %s", err)
	}

	fmt.Print(html)
}
