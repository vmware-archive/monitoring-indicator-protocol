package main

import (
	"encoding/json"
	"errors"
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
	outputFormat := flag.String("format", "bookbinder", "output format [htlm,bookbinder,prometheus-alerts,grafana]")
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

	// This will print data based on what the user provided, so it may be unsafe or contain unsanitized HTML.
	fmt.Print(output)
}

func parseDocument(format string, metadata string, filePath string) (string, error) {
	switch format {
	case "bookbinder":
		bookbinder, err := docs.DocumentToBookbinder(getDocument(filePath, indicator.SkipMetadataInterpolation))
		if err != nil {
			return "", errors.New("could not parse specified document as bookbinder")
		}
		return bookbinder, nil
	case "html":
		html, err := docs.DocumentToHTML(getDocument(filePath, indicator.SkipMetadataInterpolation))
		if err != nil {
			return "", errors.New("could not parse specified document as HTML")
		}
		return html, nil

	case "grafana":
		grafanaDashboard, err := grafana_dashboard.DocumentToDashboard(getDocument(filePath,
			indicator.OverrideMetadata(indicator.ParseMetadata(metadata))))
		if err != nil {
			return "", errors.New("could not parse specified document as Grafana dashboard")
		}
		yamlOutput, err := json.Marshal(grafanaDashboard)
		return string(yamlOutput), err
	case "prometheus-alerts":
		yamlOutput, err := yaml.Marshal(prometheus_alerts.AlertDocumentFrom(getDocument(filePath,
			indicator.OverrideMetadata(indicator.ParseMetadata(metadata)))))
		if err != nil {
			return "", errors.New("could not parse specified document as prometheus alert")
		}
		return string(yamlOutput), nil

	default:
		return "", errors.New("could not parse specified document; specified format not supported")
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
