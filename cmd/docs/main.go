package main

import (
	"fmt"
	"log"
	"os"

	"code.cloudfoundry.org/indicators/pkg/docs"
	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func main() {
	document, err := indicator.ReadFile(os.Args[1], indicator.SkipMetadataInterpolation)
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}

	html, err := docs.DocumentToHTML(document)
	if err != nil {
		log.Fatalf("cannot render document: %s", err)
	}

	fmt.Print(html)
}
