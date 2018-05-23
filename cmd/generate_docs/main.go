package main

import (
	"os"
	"io/ioutil"
	"log"
	"fmt"

	"github.com/cloudfoundry-incubator/event-producer/pkg/indicator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/docs"
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
