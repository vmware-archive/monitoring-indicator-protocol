package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/exporter"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"gopkg.in/yaml.v2"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	outputDirectory := flag.String("output-directory", "", "Indicator output-directory URI")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Server TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")
	prometheusURI := flag.String("prometheus", "", "URI of a Prometheus server instance")

	flag.Parse()

	tlsConfig, err := mtls.NewClientConfig(*clientPEM, *clientKey, *rootCACert, *serverCommonName)
	if err != nil {
		log.Fatal("failed to create mTLS HTTP client")
	}

	registryHttpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   tlsConfig,
		},
	}

	apiClient := registry.NewAPIClient(*registryURI, registryHttpClient)
	prometheusClient := &prometheusClient{
		prometheusURI: *prometheusURI,
		httpClient:    &http.Client{},
	}

	controller := exporter.NewController(exporter.ControllerConfig{
		RegistryAPIClient: apiClient,
		OutputDirectory:   *outputDirectory,
		UpdateFrequency:   time.Minute,
		DocType:           "prometheus alert",
		Converter:         Convert,
		Reloader:          prometheusClient.Reload,
	})

	controller.Start()
}

type prometheusClient struct {
	prometheusURI string
	httpClient    *http.Client
}

func (p *prometheusClient) Reload() error {
	buffer := bytes.NewBuffer(nil)
	resp, err := p.httpClient.Post(fmt.Sprintf("%s/-/reload", p.prometheusURI), "", buffer)
	if err != nil {
		return errors.New("error reloading prometheus client")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("received %v response from prometheus: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func Convert(document v1.IndicatorDocument) (*exporter.File, error) {
	documentBytes, err := yaml.Marshal(prometheus_alerts.AlertDocumentFrom(document))
	if err != nil {
		return nil, err
	}

	return &exporter.File{
		Name:     prometheus_alerts.AlertDocumentFilename(documentBytes, document.Spec.Product.Name),
		Contents: documentBytes,
	}, nil
}
