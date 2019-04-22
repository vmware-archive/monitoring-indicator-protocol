package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_uaa_client"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func main() {
	registryURI := flag.String("registry-uri", "", "URI of a registry instance")
	prometheusURI := flag.String("prometheus-uri", "", "URI of a Prometheus instance")
	updateInterval := flag.Duration("interval", 1*time.Minute, "Status update interval")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Server TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")
	insecure := flag.Bool("k", false, "skips ssl verification (insecure)")
	uaaHost := flag.String("uaa-uri", "", "URI of a UAA instance)")
	uaaClientID := flag.String("uaa-client-id", "", "UAA client ID with access to the Prometheus instance")
	uaaClientSecret := flag.String("uaa-client-secret", "", "UAA client secret")
	flag.Parse()

	checkRequiredFlagsArePresent(*registryURI, *prometheusURI, *clientPEM, *clientKey, *rootCACert, *uaaHost, *uaaClientID, *uaaClientSecret)

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

	registryClient := registry.NewAPIClient(*registryURI, registryHttpClient)

	tokenFetcher := prometheus_uaa_client.NewUAATokenFetcher(prometheus_uaa_client.UAAClientConfig{
		Insecure:        *insecure,
		UAAHost:         *uaaHost,
		UAAClientID:     *uaaClientID,
		UAAClientSecret: *uaaClientSecret,
		Timeout:         0,
	})
	prometheusClient, err := prometheus_uaa_client.Build(*prometheusURI, tokenFetcher.GetClientToken, *insecure)

	statusController := indicator_status.NewStatusController(
		registryClient,
		registryClient,
		prometheusClient,
		*updateInterval,
	)

	statusController.Start()
}

func checkRequiredFlagsArePresent(
	registryURI string,
	prometheusURI string,
	clientPEM string,
	clientKey string,
	rootCACert string,
	uaaHost string,
	uaaClientID string,
	uaaClientSecret string,
) {
	stdErr := log.New(os.Stderr, "", 0)

	exitOnEmpty := func(v string, name string) {
		if v == "" {
			stdErr.Printf("%s is required\n\n", name)
			stdErr.Printf("Usage:\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	exitOnEmpty(registryURI, "registry-uri")
	exitOnEmpty(prometheusURI, "prometheus-uri")
	exitOnEmpty(clientPEM, "tls-pem-path")
	exitOnEmpty(clientKey, "tls-key-path")
	exitOnEmpty(rootCACert, "tls-root-ca-pem")
	exitOnEmpty(uaaHost, "uaa-host")
	exitOnEmpty(uaaClientID, "uaa-client-id")
	exitOnEmpty(uaaClientSecret, "uaa-client-secret")
}
