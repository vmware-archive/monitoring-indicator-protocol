package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
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

	startMetricsEndpoint()

	agent := registry.Agent{
		DocumentFinder:  registry.DocumentFinder{Glob: *documentsGlob},
		RegistryURI:    *registryURI,
		DeploymentName: *deploymentName,
		IntervalTime:   *intervalTime,
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
