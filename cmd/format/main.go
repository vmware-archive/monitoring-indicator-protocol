package main

import (
	"flag"
	"fmt"
	"log"

	"code.cloudfoundry.org/indicators/pkg/docs"
	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func main() {
	output := flag.String("format", "bookbinder", "output format [bookbinder,html,grafana]")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("only one file argument allowed\n")
	}

	document, err := indicator.ReadFile(args[0], indicator.SkipMetadataInterpolation)
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}
	var text string

	switch *output {
	case "bookbinder":
		text, err = docs.DocumentToHTML(document)
		if err != nil {
			log.Fatalf("cannot render document: %s\n", err)
		}
	}

	fmt.Print(text)
}
