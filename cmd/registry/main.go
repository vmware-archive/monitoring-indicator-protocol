package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"code.cloudfoundry.org/indicators/pkg/registry"
)

func main() {
	port := flag.Int("port", 443, "Port to expose registration endpoints")
	serverPEM := flag.String("tls-pem-path", "", "Server TLS public cert pem path")
	serverKey := flag.String("tls-key-path", "", "Server TLS private key path")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs.")
	expiration := flag.Duration("indicator-expiration", 120*time.Minute, "Document expiration duration")
	flag.Parse()

	address := fmt.Sprintf(":%d", *port)

	config := registry.WebServerConfig{
		Address:       address,
		ServerPEMPath: *serverPEM,
		ServerKeyPath: *serverKey,
		RootCAPath:    *rootCACert,
		DocumentStore: registry.NewDocumentStore(*expiration),
	}

	start, stop, err := registry.NewWebServer(config)

	if err != nil {
		log.Fatalf("failed to create server: %s\n", err)
	}
	defer stop()

	err = start()
	if err != nil {
		log.Fatalf("failed to create server: %s\n", err)
	}
}
