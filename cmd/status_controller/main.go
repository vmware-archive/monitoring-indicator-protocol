package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_oauth_client"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func main() {
	registryURI := flag.String("registry-uri", "", "URI of a registry instance")
	prometheusURI := flag.String("prometheus-uri", "", "URI of a Prometheus instance")
	updateInterval := flag.Duration("interval", 1*time.Minute, "Status update interval")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Client TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")
	insecure := flag.Bool("k", false, "skips ssl verification (insecure)")
	oauthHost := flag.String("oauth-server", "", "URI of a OAuth authentication server)")
	oauthClientID := flag.String("oauth-client-id", "", "OAuth client ID with access to the Prometheus instance")
	oauthClientSecret := flag.String("oauth-client-secret", "", "OAuth client secret")
	flag.Parse()

	checkRequiredFlagsArePresent(*registryURI, *prometheusURI, *clientPEM, *clientKey, *rootCACert, *oauthHost, *oauthClientID, *oauthClientSecret)

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

	registryClient := registry.NewAPIClient(*registryURI, registryHttpClient)

	tokenFetcher := prometheus_oauth_client.NewTokenFetcher(prometheus_oauth_client.OAuthClientConfig{
		Insecure:          *insecure,
		OAuthServer:       *oauthHost,
		OAuthClientID:     *oauthClientID,
		OAuthClientSecret: *oauthClientSecret,
		Timeout:           0,
	})
	prometheusClient, err := prometheus_oauth_client.Build(*prometheusURI, tokenFetcher.GetClientToken, *insecure)

	if err != nil {
		log.Fatal("failed to create Prometheus client")
	}

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
	oauthHost string,
	oauthClientID string,
	oauthClientSecret string,
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
	exitOnEmpty(oauthHost, "oauth-server")
	exitOnEmpty(oauthClientID, "oauth-client-id")
	exitOnEmpty(oauthClientSecret, "oauth-client-secret")
}
