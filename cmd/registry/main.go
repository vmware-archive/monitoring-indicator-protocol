package main

import (
	"code.cloudfoundry.org/indicators/pkg/registry"
	"flag"
	"fmt"
	"log"
	"time"
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
		Address:    address,
		ServerPEM:  *serverPEM,
		ServerKey:  *serverKey,
		RootCACert: *rootCACert,
		Expiration: *expiration,
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
