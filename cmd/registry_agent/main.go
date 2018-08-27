package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"code.cloudfoundry.org/cf-indicators/pkg/registry"
)

func main() {
	indicatorsPath := flag.String("indicators-path", "./", "Path to a directory containing indicator files")
	registryURI := flag.String("registry", "", "URI of a registry instance")
	deploymentName := flag.String("deployment", "", "The name of the deployment")
	productName := flag.String("product", "", "The name of the product")
	intervalTime := flag.Duration("interval", 5*time.Minute, "The send interval")
	flag.Parse()

	document, err := ioutil.ReadFile(*indicatorsPath)
	if err != nil {
		log.Fatalf("could not read indicator document: %s/n", err)
	}

	agent := registry.Agent{
		IndicatorsDocument: document,
		RegistryURI:        *registryURI,
		DeploymentName:     *deploymentName,
		ProductName:        *productName,
		IntervalTime:       *intervalTime,
	}

	startMetricsEndpoint()

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
