package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/mtls"
	"github.com/pivotal/indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/indicator-protocol/pkg/registry"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	outputDirectory := flag.String("output-directory", "", "Indicator output-directory URI")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Server TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")

	flag.Parse()

	tlsConfig, err := mtls.NewClientConfig(*clientPEM, *clientKey, *rootCACert, *serverCommonName)
	if err != nil {
		log.Fatalf("failed to create mtls http client, %s", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   tlsConfig,
		},
	}

	c := registry.NewAPIClient(*registryURI, client)

	documents, err := c.IndicatorDocuments()
	if err != nil {
		log.Fatalf("failed to fetch indicator documents, %s", err)
	}

	writeDocuments(formatDocuments(documents), *outputDirectory)
}

func formatDocuments(documents []registry.APIV0Document) []indicator.Document {
	formattedDocuments := make([]indicator.Document, 0)
	for _, d := range documents {
		formattedDocuments = append(formattedDocuments, convertDocument(d))
	}

	return formattedDocuments
}

func convertDocument(d registry.APIV0Document) indicator.Document {
	indicators := make([]indicator.Indicator, 0)
	for _, i := range d.Indicators {
		indicators = append(indicators, convertIndicator(i))
	}

	return indicator.Document{
		Product: indicator.Product{
			Name:    d.Product.Name,
			Version: d.Product.Version,
		},
		Indicators: indicators,
	}
}

func convertIndicator(i registry.APIV0Indicator) indicator.Indicator {
	thresholds := make([]indicator.Threshold, 0)
	for _, t := range i.Thresholds {
		thresholds = append(thresholds, convertThreshold(t))
	}

	return indicator.Indicator{
		Name:          i.Name,
		PromQL:        i.PromQL,
		Thresholds:    thresholds,
		Documentation: i.Documentation,
	}
}

func convertThreshold(t registry.APIV0Threshold) indicator.Threshold {
	return indicator.Threshold{
		Level:    t.Level,
		Operator: indicator.GetComparatorFromString(t.Operator),
		Value:    t.Value,
	}
}

func writeDocuments(documents []indicator.Document, outputDirectory string) {
	for _, d := range documents {
		fileBytes, _ := yaml.Marshal(prometheus_alerts.AlertDocumentFrom(d))
		err := ioutil.WriteFile(fmt.Sprintf("%s/%s.yml", outputDirectory, d.Product.Name), fileBytes, 0644)
		if err != nil {
			log.Printf("error writing file: %s\n", err)
		}
	}
}
