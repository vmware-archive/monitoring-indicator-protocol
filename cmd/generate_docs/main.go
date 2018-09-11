package main

import (
	"os"
	"log"
	"fmt"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/docs"
)

func main() {
	document, err := indicator.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}

	html, err := docs.DocumentToHTML(document)
	if err != nil {
		log.Fatalf("cannot render document: %s", err)
	}

	fmt.Print(html)
}
