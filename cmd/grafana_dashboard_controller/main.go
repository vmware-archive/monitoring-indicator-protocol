package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pivotal/indicator-protocol/pkg/exporter"
	"github.com/pivotal/indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/mtls"
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

	registryHttpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   tlsConfig,
		},
	}

	apiClient := registry.NewAPIClient(*registryURI, registryHttpClient)

	exporterController := exporter.NewController(exporter.ControllerConfig{
		RegistryAPIClient: apiClient,
		OutputDirectory:   *outputDirectory,
		UpdateFrequency:   time.Minute,
		DocType:           "grafana dashboard",
		Converter:         Convert,
	})

	exporterController.Start()
}

func Convert(document indicator.Document) (*exporter.File, error) {
	documentString, err := grafana_dashboard.DocumentToDashboard(document)
	if err != nil {
		return nil, err
	}
	return &exporter.File{
		Name:     fmt.Sprintf("%s_%x.json", document.Product.Name, sha1.Sum([]byte(documentString))),
		Contents: []byte(documentString),
	}, nil
}
