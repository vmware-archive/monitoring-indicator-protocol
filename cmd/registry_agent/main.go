package main

import (
	"flag"
	"time"
	"net/http"
	"fmt"
	"bytes"
	"io/ioutil"
	"log"
)

func main() {
	indicatorsPath := flag.String("indicators-path", "./", "Path to a directory containing indicator files")
	registryURI := flag.String("registry", "", "URI of a registry instance")
	deploymentName := flag.String("deployment", "", "The name of the deployment")
	intervalTime := flag.Duration("interval", 5*time.Minute, "The send interval")
	flag.Parse()

	files, err := ioutil.ReadDir(*indicatorsPath)
	if err != nil {
		log.Fatalf("could not read filepath: %s\n", err)
	}

	interval := time.NewTicker(*intervalTime)
	for {
		select {
		case <-interval.C:
			for _, fileInfo := range files {
				registry := fmt.Sprintf(*registryURI+"/v1/register?deployment=%s&product=%s", *deploymentName, fileInfo.Name())

				fileData, err := ioutil.ReadFile(*indicatorsPath + fileInfo.Name())
				if err != nil {
					log.Fatalf("could not read indicators file: %s\n", err)
				}
				body := bytes.NewBuffer(fileData)

				_, err = http.Post(registry, "text/plain", body)
				if err != nil {
					// TODO: emit failure metric
					log.Fatalf("could not make http request: %s\n", err)
				}
			}
		default:
		}
	}
}
