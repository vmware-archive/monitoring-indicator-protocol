package registry

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type Agent struct {
	IndicatorsDocuments [][]byte
	RegistryURI         string
	DeploymentName      string
	IntervalTime        time.Duration
}

var registrationCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name:        "registration_count",
	Help:        "counter of all registration attempts",
	ConstLabels: nil,
}, []string{"status"})

func init() {
	prometheus.MustRegister(registrationCount)
}

func (a Agent) Start() {
	a.registerIndicatorDocuments()

	interval := time.NewTicker(a.IntervalTime)
	for {
		select {
		case <-interval.C:
			for _, d := range a.IndicatorsDocuments {
				a.registerIndicatorDocument(d)
			}
		default:
		}
	}
}

func (a Agent) registerIndicatorDocuments() {
	for _, d := range a.IndicatorsDocuments {
		a.registerIndicatorDocument(d)
	}
}

func (a Agent) registerIndicatorDocument(indicatorsDocument []byte) {

	registry := fmt.Sprintf(a.RegistryURI+"/v1/register?deployment=%s", a.DeploymentName)

	body := bytes.NewBuffer(indicatorsDocument)

	resp, err := http.Post(registry, "text/plain", body)
	if err != nil {
		registrationCount.WithLabelValues("err").Inc()
		log.Printf("could not make http request: %s\n", err)
	} else {
		registrationCount.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()
	}
}
