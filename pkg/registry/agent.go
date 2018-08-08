package registry

import (
	"time"
	"io/ioutil"
	"log"
	"fmt"
	"bytes"
	"net/http"
)

type Agent struct {
	IndicatorsPath string
	RegistryURI    string
	DeploymentName string
	IntervalTime   time.Duration
}

func (a Agent) Start() {
	files, err := ioutil.ReadDir(a.IndicatorsPath)
	if err != nil {
		log.Fatalf("could not read filepath: %s\n", err)
	}

	interval := time.NewTicker(a.IntervalTime)
	for {
		select {
		case <-interval.C:
			for _, fileInfo := range files {
				registry := fmt.Sprintf(a.RegistryURI+"/v1/register?deployment=%s&product=%s", a.DeploymentName, fileInfo.Name())

				fileData, err := ioutil.ReadFile(a.IndicatorsPath + fileInfo.Name())
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