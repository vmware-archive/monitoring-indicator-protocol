package main

import (
	"flag"
	"time"
	"code.cloudfoundry.org/cf-indicators/pkg/registry"
)

func main() {
	indicatorsPath := flag.String("indicators-path", "./", "Path to a directory containing indicator files")
	registryURI := flag.String("registry", "", "URI of a registry instance")
	deploymentName := flag.String("deployment", "", "The name of the deployment")
	intervalTime := flag.Duration("interval", 5*time.Minute, "The send interval")
	flag.Parse()

	agent := registry.Agent{
		IndicatorsPath: *indicatorsPath,
		RegistryURI:    *registryURI,
		DeploymentName: *deploymentName,
		IntervalTime:   *intervalTime,
	}

	agent.Start()
}
