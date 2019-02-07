package main

import (
	"flag"
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"log"
	"time"

	"github.com/pivotal/indicator-protocol/pkg/configuration"
	"github.com/pivotal/indicator-protocol/pkg/registry"
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
		}
	}
}

func upsertFromConfig(configFile string, store *registry.DocumentStore) {
	patches, documents, err := configuration.Read(configFile, getRealRepository)
	if err != nil {
		log.Fatalf("failed to read configuration file: %s\n", err)
	}
	for _, p := range patches {
		store.UpsertPatches(p)
	}
	for _, d := range documents {
		store.UpsertDocument(d)
	}
}

func getRealRepository(s configuration.Source) (*git.Repository, error) {
	storage := memory.NewStorage()
	var auth transport.AuthMethod = nil
	if s.Token != "" {
		auth = &http.BasicAuth{
			Username: "github",
			Password: s.Token,
		}
	}
	repoURL := s.Repository
	r, err := git.Clone(storage, nil, &git.CloneOptions{
		Auth: auth,
		URL:  repoURL,
	})
	return r, err
}
