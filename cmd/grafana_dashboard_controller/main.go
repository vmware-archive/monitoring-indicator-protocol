package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/exporter"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/tls_config"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func main() {
	indicatorTypeInput := flag.String("indicator-type", "all", "Only generate dashboards from specific indicators [sli, kpi, other, all]")
	registryURI := flag.String("registry", "", "URI of a registry proxy instance")
	outputDirectory := flag.String("output-directory", "", "Indicator output-directory URI")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Client TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")

	flag.Parse()

	tlsConfig, err := tls_config.NewClientConfig(*clientPEM, *clientKey, *rootCACert, *serverCommonName)
	if err != nil {
		log.Fatal("construction of grafana_dashboard_controller failed, could not create mTLS HTTP client")
	}

	indicatorType := v1.IndicatorTypeFromString(*indicatorTypeInput)

	if indicatorType == v1.UndefinedType && *indicatorTypeInput != "all" {
		log.Fatal("Invalid `indicator-type`, must be one of [sli, kpi, other, all]")
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
		Converter: func(document v1.IndicatorDocument) (file *exporter.File, e error) {
			grafanaDashboard, err := grafana_dashboard.DocumentToDashboard(document, indicatorType)
			if err != nil {
				return nil, err
			}
			if grafanaDashboard == nil {
				return nil, nil
			}
			documentBytes, err := json.Marshal(grafanaDashboard)
			if err != nil {
				return nil, err
			}
			return &exporter.File{
				Name:     grafana_dashboard.DashboardFilename(documentBytes, document.Spec.Product.Name),
				Contents: documentBytes,
			}, nil
		},
	})

	exporterController.Start()
}
