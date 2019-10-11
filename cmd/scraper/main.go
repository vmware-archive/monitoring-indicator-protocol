package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ghodss/yaml"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/scraper"
)

// TODO: dockerize?

func main() {
	interval := flag.Duration("interval", 60*time.Second, "Scrape interval")
	localKey := flag.String("local-key-path", "", "Local registry client key path")
	localPem := flag.String("local-pem-path", "", "Local registry client cert path")
	localCaPem := flag.String("local-root-ca-pem", "", "Local registry ca path")
	localAddr := flag.String("local-registry-addr", "", "Local registry URL")
	localCommonName := flag.String("local-server-cn", "", "Local registry server name")
	remoteScrapeConfig := flag.String("remote-scrape-configs-path", "", "Remote scrape configurations YAML file path")
	flag.Parse()

	localTlsClientConfig, err := mtls.NewClientConfig(*localPem, *localKey, *localCaPem, *localCommonName)
	if err != nil {
		log.Fatalf("Error with creating mTLS local client config: %s", err)
	}

	localApiClient := registry.NewAPIClient(*localAddr, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: localTlsClientConfig,
		},
	})

	remoteScrapeConfigBytes, err := ioutil.ReadFile(*remoteScrapeConfig)
	var scrapeConfigs []scraper.RemoteScrapeConfig
	err = yaml.Unmarshal(remoteScrapeConfigBytes, &scrapeConfigs)
	if err != nil {
		log.Fatalf("Could not read remote scrape config file %s", err)
	}

	remoteApiClients, errs := scraper.MakeApiClients(scrapeConfigs)
	if len(errs) != 0 {
		errString := "Errors creating remote API clients: \n"
		for _, err := range errs {
			errString += " - " + err.Error() + "\n"
		}
		log.Fatalf(errString)
	}

	scraper.RunLoop(*interval, remoteApiClients, localApiClient)
}

