package main

import (
	"flag"
	"fmt"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/cf_registry_proxy"
	"log"
	"net/http"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/gorilla/mux"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Server TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")
	capiAddress := flag.String("capi-address", "", "Cloud Controller API Address")
	port := flag.Int("port", 8080, "HTTP Port to listen on")

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

	router := mux.NewRouter()
	router.HandleFunc("/indicator-documents", cf_registry_proxy.IndicatorDocumentsHandler(apiClient, *capiAddress)).Queries("service_instance_guid", "{service_instance_guid}")
	http.ListenAndServe(fmt.Sprintf(":%d", port), router)

	_, err = apiClient.IndicatorDocuments()


	c := &cfclient.Config{
		ApiAddress:   *capiAddress,
	}

	client, _ := cfclient.NewClient(c)
	client.ServiceInstanceByGuid()

}


