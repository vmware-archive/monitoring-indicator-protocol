package main

import (
	"code.cloudfoundry.org/indicators/pkg/mtls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"code.cloudfoundry.org/indicators/pkg/registry"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	intervalTime := flag.Duration("interval", 5*time.Minute, "The send interval")
	documentsGlob := flag.String("documents-glob", "/var/vcap/jobs/*/indicators.yml", "Glob path of indicator files")

	clientPEM := flag.String("tls-pem-path", "", "Server TLS public cert pem path")
	clientKey := flag.String("tls-key-path", "", "Server TLS private key path")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs.")
	flag.Parse()

	startMetricsEndpoint()

	client, err := mtls.NewClient(*clientPEM, *clientKey, *rootCACert)
	if err != nil {
		log.Fatalf("failed to create mtls http client, %s", err)
	}

	agent := registry.Agent{
		DocumentFinder: registry.DocumentFinder{Glob: *documentsGlob},
		RegistryURI:    *registryURI,
		IntervalTime:   *intervalTime,
		Client:         client,
	}
	agent.Start()
}

func startMetricsEndpoint() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 0))
	if err != nil {
		log.Printf("unable to start monitor endpoint: %s", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	log.Printf("starting monitor endpoint on http://%s/metrics\n", lis.Addr().String())
	go func() {
		err = http.Serve(lis, mux)
		log.Printf("error starting the monitor server: %s", err)
	}()
}
