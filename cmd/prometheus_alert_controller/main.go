package main

import (
	"flag"

	"github.com/pivotal/indicator-protocol/pkg/prometheus_alerts"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	outputDirectory := flag.String("output-directory", "", "Indicator output-directory URI")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Server TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")

	flag.Parse()

	c := prometheus_alerts.ControllerConfig{
		RegistryURI:       *registryURI,
		TLSPEMPath:        *clientPEM,
		TLSKeyPath:        *clientKey,
		TLSRootCACertPath: *rootCACert,
		TLSServerCN:       *serverCommonName,
		OutputDirectory:   *outputDirectory,
	}

	controller := prometheus_alerts.NewController(c)
	controller.Update()
}
