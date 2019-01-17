package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"code.cloudfoundry.org/indicators/pkg/configuration"
	"code.cloudfoundry.org/indicators/pkg/registry"
)

func main() {
	port := flag.Int("port", 443, "Port to expose registration endpoints")
	serverPEM := flag.String("tls-pem-path", "", "Server TLS public cert pem path")
	serverKey := flag.String("tls-key-path", "", "Server TLS private key path")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs.")
	expiration := flag.Duration("indicator-expiration", 120*time.Minute, "Document expiration duration")
	configFile := flag.String("config", "", "Configuration yaml for patch and document sources")

	flag.Parse()

	address := fmt.Sprintf(":%d", *port)

	store := registry.NewDocumentStore(*expiration)

	if *configFile != "" {
		upsertFromConfig(*configFile, store)
		go readConfigEachMinute(*configFile, store)
	}

	config := registry.WebServerConfig{
		Address:       address,
		ServerPEMPath: *serverPEM,
		ServerKeyPath: *serverKey,
		RootCAPath:    *rootCACert,
		DocumentStore: store,
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

func readConfigEachMinute(configFile string, store *registry.DocumentStore) {
	timer := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-timer.C:
			upsertFromConfig(configFile, store)
		default:
		}
	}
}

func upsertFromConfig(configFile string, store *registry.DocumentStore) {
	patches, documents, err := configuration.Read(configFile)
	if err != nil {
		log.Fatalf("failed to read configuration file: %s\n", err)
	}
	for _, p := range patches {
		store.UpsertPatch(p)
	}
	for _, d := range documents {
		store.UpsertDocument(d)
	}
}
