package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"

	"github.com/pivotal/indicator-protocol/pkg/docs"
	"github.com/pivotal/indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/prometheus_alerts"
)

func main() {
	outputFormat := flag.String("format", "bookbinder", "output format [bookbinder,prometheus-alerts,grafana]")
	metadata := flag.String("metadata", "", "metadata to override (e.g. --metadata deployment=my-test-deployment,source_id=metric-forwarder)")
	indicatorsFilePath := flag.String("indicators", "", "indicators YAML file path")

	flag.Parse()

	if len(*indicatorsFilePath) == 0 {
		log.Fatalf("-indicators flag is required")
	}

	output, err := parseDocument(*outputFormat, *metadata, *indicatorsFilePath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(output)

}

func parseDocument(format string, metadata string, filePath string) (string, error) {
	switch format {
	case "bookbinder":
		return docs.DocumentToBookbinder(getDocument(filePath, indicator.SkipMetadataInterpolation))
	case "html":
		return docs.DocumentToHTML(getDocument(filePath, indicator.SkipMetadataInterpolation))
	case "grafana":
		return grafana_dashboard.DocumentToDashboard(getDocument(filePath,
			indicator.OverrideMetadata(indicator.ParseMetadata(metadata))))
	case "prometheus-alerts":
		yamlOutput, err := yaml.Marshal(prometheus_alerts.AlertDocumentFrom(getDocument(filePath,
			indicator.OverrideMetadata(indicator.ParseMetadata(metadata)))))
		return string(yamlOutput), err

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
