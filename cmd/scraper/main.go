package main

import (
	"flag"
	"log"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
)

func main() {
	interval := flag.Duration("interval", 60*time.Second, "TODO")
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


	_ = localAddr
	_ = remoteAddr
	_ = remoteTlsClientConfig
	_ = localTlsClientConfig

	// 	On a loop: poll fromServer, add metadata, send to toServer
	ticker := time.NewTicker(*interval)

	for {
		select {
		case <-ticker.C:
			// 	Send request to remote, etc.

		}
	}
}
