package registry

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"io/ioutil"
	"net/http"
)

type Agent struct {
	IndicatorsDocument string
	RegistryURI    string
	DeploymentName string
	ProductName    string
	IntervalTime   time.Duration
}


func (a Agent) Start() {
	file, err := ioutil.ReadFile(a.IndicatorsDocument)
	if err != nil {
		log.Fatalf("could not read file: %s\n", err)
	}


	interval := time.NewTicker(a.IntervalTime)
	for {
		select {
		case <-interval.C:
				registry := fmt.Sprintf(a.RegistryURI+"/v1/register?deployment=%s&product=%s", a.DeploymentName, a.ProductName)

				body := bytes.NewBuffer(file)

				_, err = http.Post(registry, "text/plain", body)
				if err != nil {
					// TODO: emit failure metric
					log.Fatalf("could not make http request: %s\n", err)
				}
		default:
		}
	}
}
