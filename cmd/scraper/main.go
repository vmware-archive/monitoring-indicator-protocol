package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func main() {
	interval := flag.Duration("interval", 60*time.Second, "TODO")
	sourceName := flag.String("source-name", "", "TODO")
	localKey := flag.String("local-key-path", "", "TODO")
	remoteKey := flag.String("remote-key-path", "", "TODO")
	localPem := flag.String("local-pem-path", "", "TODO")
	remotePem := flag.String("remote-pem-path", "", "TODO")
	localCaPem := flag.String("local-root-ca-pem", "", "TODO")
	remoteCaPem := flag.String("remote-root-ca-pem", "", "TODO")
	localAddr := flag.String("local-registry-addr", "", "TODO")
	remoteAddr := flag.String("remote-registry-addr", "", "TODO")
	localCommonName := flag.String("local-server-cn", "", "TODO")
	remoteCommonName := flag.String("remote-server-cn", "", "TODO")
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
