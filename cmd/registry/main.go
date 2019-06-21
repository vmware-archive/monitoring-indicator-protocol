package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/configuration"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func main() {
	port := flag.Int("port", 10568, "Port to expose registration endpoints")
	host := flag.String("host", "localhost", "Host to bind to for registration endpoints")
	expiration := flag.Duration("indicator-expiration", 120*time.Minute, "Document expiration duration")
	configFile := flag.String("config", "", "Configuration yaml for patch and document sources")

	flag.Parse()

	address := fmt.Sprintf("%s:%d", *host, *port)

	store := registry.NewDocumentStore(*expiration, time.Now)

	if *configFile != "" {
		upsertFromConfig(*configFile, store)
		go readConfigEachMinute(*configFile, store)
	}

	config := registry.WebServerConfig{
		Address:       address,
		DocumentStore: store,
		StatusStore:   status_store.New(time.Now),
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
	timer := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-timer.C:
			upsertFromConfig(configFile, store)
		}
	}
}

func upsertFromConfig(configFile string, store *registry.DocumentStore) {
	sources, err := configuration.ParseSourcesFile(configFile)
	if err != nil {
		log.Fatalf("failed to parse configuration file: %s\n", err)
	}
	patches, documents := configuration.Read(sources, getRealRepository)

	for _, p := range patches {
		store.UpsertPatches(p)
	}
	for _, d := range documents {
		store.UpsertDocument(d)
	}
}

func getRealRepository(s configuration.Source) (*git.Repository, error) {
	storage := memory.NewStorage()
	auth := getAuth(s)
	repoURL := s.Repository
	r, err := git.Clone(storage, nil, &git.CloneOptions{
		Auth: auth,
		URL:  repoURL,
	})
	return r, err
}

func getAuth(s configuration.Source) transport.AuthMethod {
	if s.Token != "" {
		return &http.BasicAuth{
			Username: "github",
			Password: s.Token,
		}
	}
	if s.Key != "" {
		signer, _ := ssh.ParsePrivateKey([]byte(s.Key))
		helper := ssh2.HostKeyCallbackHelper{
			HostKeyCallback: noopHostkeyCallback,
		}
		return &ssh2.PublicKeys{User: "git", Signer: signer, HostKeyCallbackHelper: helper}
	}
	return nil
}

func noopHostkeyCallback(_ string, _ net.Addr, _ ssh.PublicKey) error {
	return nil
}
