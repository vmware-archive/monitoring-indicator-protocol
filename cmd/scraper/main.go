package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

// TODO: dockerize?

func main() {
	interval := flag.Duration("interval", 60*time.Second, "Scrape interval")
	sourceName := flag.String("source-name", "", "Additional metadata value that will appear in the scraped documents (e.g. foundation name)")
	localKey := flag.String("local-key-path", "", "Local registry client key path")
	remoteKey := flag.String("remote-key-path", "", "Remote registry client key path")
	localPem := flag.String("local-pem-path", "", "Local registry client cert path")
	remotePem := flag.String("remote-pem-path", "", "Remote registry client cert path")
	localCaPem := flag.String("local-root-ca-pem", "", "Local registry ca path")
	remoteCaPem := flag.String("remote-root-ca-pem", "", "Remote registry ca path")
	localAddr := flag.String("local-registry-addr", "", "Local registry URL")
	remoteAddr := flag.String("remote-registry-addr", "", "Remote registry URL")
	localCommonName := flag.String("local-server-cn", "", "Local registry server name")
	remoteCommonName := flag.String("remote-server-cn", "", "Remote registry server name")
	flag.Parse()

	remoteTlsClientConfig, err := mtls.NewClientConfig(*remotePem, *remoteKey, *remoteCaPem, *remoteCommonName)
	if err != nil {
		log.Fatalf("Error with creating mTLS client config: %s", err)
	}

	localTlsClientConfig, err := mtls.NewClientConfig(*localPem, *localKey, *localCaPem, *localCommonName)
	if err != nil {
		log.Fatalf("Error with creating mTLS client config: %s", err)
	}

	localApiClient := registry.NewAPIClient(*localAddr, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: localTlsClientConfig,
		},
	})

	remoteApiClient := registry.NewAPIClient(*remoteAddr, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: remoteTlsClientConfig,
		},
	})

	// 	On a loop: poll fromServer, add metadata, send to toServer
	ticker := time.NewTicker(*interval)
	for {
		select {
		case <-ticker.C:
			// 	Send request to remote, etc.
			documents, err := remoteApiClient.IndicatorDocuments()
			if err != nil {
				log.Printf("could not retrieve documents: %s", err)
				continue
			}
			for _, document := range documents {
				document.Metadata.Labels["source_name"] = *sourceName
				err := localApiClient.Register(registry.ToIndicatorDocument(document))
				if err != nil {
					log.Printf("could not post documents: %s", err)
					continue
				}
			}
		}
	}
}
