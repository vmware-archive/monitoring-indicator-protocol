package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/docs"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
)

func main() {
	l := log.New(os.Stderr, "", 0)
	outputFormat := flag.String("format", "bookbinder", "output format [bookbinder,prometheus-alerts,grafana]")
	metadata := flag.String("metadata", "", "metadata to override (e.g. --metadata deployment=my-test-deployment,source_id=metric-forwarder)")
	indicatorsFilePath := flag.String("indicators", "", "indicators YAML file path")

	flag.Parse()

	if len(*indicatorsFilePath) == 0 {
		l.Fatalf("-indicators flag is required")
	}

	output, err := parseDocument(*outputFormat, *metadata, *indicatorsFilePath)
	if err != nil {
		l.Fatal(err)
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
		yamlOutput, err := json.Marshal(grafana_dashboard.DocumentToDashboard(getDocument(filePath,
			indicator.OverrideMetadata(indicator.ParseMetadata(metadata)))))
		return string(yamlOutput), err
	case "prometheus-alerts":
		yamlOutput, err := yaml.Marshal(prometheus_alerts.AlertDocumentFrom(getDocument(filePath,
			indicator.OverrideMetadata(indicator.ParseMetadata(metadata)))))
		return string(yamlOutput), err

	default:
		return "", fmt.Errorf(`format "%s" not supported`, format)
	}
}

func getDocument(docPath string, opts ...indicator.ReadOpt) indicator.Document {
	l := log.New(os.Stderr, "", 0)
	document, err := indicator.ReadFile(docPath, opts...)
	if err != nil {
		l.Fatal(err)
	}

	return document
}
