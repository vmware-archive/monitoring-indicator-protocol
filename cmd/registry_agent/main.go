package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"code.cloudfoundry.org/cf-indicators/pkg/registry"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	deploymentName := flag.String("deployment", "", "The name of the deployment")
	intervalTime := flag.Duration("interval", 5*time.Minute, "The send interval")
	documentsGlob := flag.String("documents-glob", "/var/vcap/jobs/*/indicators.yml", "Glob path of indicator files")
	flag.Parse()

	documentPaths, err := filepath.Glob(*documentsGlob)
	if err != nil {
		log.Fatalf("could not read glob indicator documents: %s/n", err)
	}

	documents := make([][]byte, 0)
	for _, path := range documentPaths {
		document, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("could not read indicator document: %s/n", err)
		}

		documents = append(documents, document)
	}

	startMetricsEndpoint()

	agent := registry.Agent{
		IndicatorsDocuments: documents,
		RegistryURI:         *registryURI,
		DeploymentName:      *deploymentName,
		IntervalTime:        *intervalTime,
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
		log.Printf("error starting the monitor server: %s", http.Serve(lis, mux))
	}()
}
