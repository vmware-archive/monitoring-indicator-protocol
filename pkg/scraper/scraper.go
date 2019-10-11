package scraper

import (
	"fmt"
	"log"
	"net/http"
	"time"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

type RemoteScrapeConfig struct {
	SourceName   string `json:"scraper_source_name"`
	ServerName   string `json:"scraper_remote_server_name"`
	RegistryAddr string `json:"scraper_remote_registry_address"`
	// The actual bytes of a cert, NOT the file that the cert is in
	CaCert      []byte      `json:"scraper_remote_ca_cert"`
	ClientCreds ClientCreds `json:"scraper_remote_client_cred"`
}

type ClientCreds struct {
	// The actual bytes of a key, NOT the file that the cert is in
	ClientKey []byte `json:"cert_pem"`
	// The actual bytes of a cert, NOT the file that the cert is in
	ClientCert []byte `json:"private_key_pem"`
}

type RemoteFoundationApiClient interface {
	ForwardDocumentsTo(destinationApiClient DocumentRegistrar)
}

type DocumentRegistrar interface {
	Register(document v1.IndicatorDocument) error
}

// Wraps an APIClient for scraping remote foundations, with a source name to add to metadata
type remoteFoundationApiClient struct {
	ApiClient  *registry.RegistryApiClient
	SourceName string
}

func RunLoop(interval time.Duration, remoteApiClients []RemoteFoundationApiClient, localApiClient DocumentRegistrar) func() {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				break
			case <-ticker.C:
				for _, remoteClient := range remoteApiClients {
					go remoteClient.ForwardDocumentsTo(localApiClient)
				}
			}
		}
	}()

	return func() {
		quit <- struct{}{}
	}
}

// Scrapes documents from the "origin" client, and then writes them to the destination API client
func (originClient remoteFoundationApiClient) ForwardDocumentsTo(destinationApiClient DocumentRegistrar) {
	documents, err := originClient.ApiClient.IndicatorDocuments()
	if err != nil {
		log.Printf("could not retrieve documents: %s", err)
		return
	}
	for _, document := range documents {
		document.Metadata.Labels["source_name"] = originClient.SourceName
		err := destinationApiClient.Register(registry.ToIndicatorDocument(document))
		if err != nil {
			log.Printf("could not post documents: %s", err)
			return
		}
	}
}

func MakeApiClients(remoteConfigs []RemoteScrapeConfig) ([]RemoteFoundationApiClient, []error) {
	apiClients := make([]RemoteFoundationApiClient, len(remoteConfigs))
	errors := make([]error, 0)

	for index, config := range remoteConfigs {
		remoteTlsClientConfig, err := mtls.NewClientConfigFromValues(
			config.ClientCreds.ClientCert,
			config.ClientCreds.ClientKey,
			config.CaCert,
			config.ServerName)
		if err != nil {
			errors = append(errors, fmt.Errorf("error with creating mTLS remote client config: %s", err))
		}
		remoteApiClient := registry.NewAPIClient(config.RegistryAddr, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: remoteTlsClientConfig,
			},
		})
		apiClients[index] = remoteFoundationApiClient{
			ApiClient:  remoteApiClient,
			SourceName: config.SourceName,
		}
	}

	return apiClients, errors
}
