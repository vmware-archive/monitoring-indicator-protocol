package main

import (
	"flag"
	"fmt"
	"log"
	
	"code.cloudfoundry.org/indicators/pkg/docs"
	"code.cloudfoundry.org/indicators/pkg/grafana_dashboard"
	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func main() {
	outputFormat := flag.String("format", "bookbinder", "output format [bookbinder,prometheus-alerts]")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("only one file argument allowed\n")
	}

	filePath := args[0]

	output, err := parseDocument(*outputFormat, filePath)
	if len(args) != 1 {
		log.Fatal(err)
	}

	fmt.Print(output)

}

func parseDocument(format string, filePath string) (string, error) {
	switch format {
	case "bookbinder":
		s, e := docs.DocumentToHTML(getDocument(filePath, indicator.SkipMetadataInterpolation))
		return string(s), e
	case "grafana":
		return grafana_dashboard.DocumentToDashboard(getDocument(filePath))

	//case "prometheus-alerts":
	//	yamlOutput, err := yaml.Marshal(prometheus_alerts.AlertDocumentFrom(getDocument(filePath)))
	//	return string(yamlOutput), err

	default:
		return "", fmt.Errorf(`format "%s" not supported`, format)
	}
}

func getDocument(docPath string, opts ...indicator.ReadOpt) indicator.Document {
	document, err := indicator.ReadFile(docPath, opts...)
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}

	return document
}
