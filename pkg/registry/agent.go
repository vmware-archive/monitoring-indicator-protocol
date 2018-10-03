package registry

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type document []byte

type Agent struct {
	RegistryURI    string
	IntervalTime   time.Duration
	DocumentFinder DocumentFinder
	Client         *http.Client
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
			a.registerIndicatorDocuments()
		}
	}
}

func (a Agent) registerIndicatorDocuments() {
	documents, err := a.DocumentFinder.FindAll()

	if err != nil {
		registrationCount.WithLabelValues("err").Inc()
		log.Printf("could not find documents: %s\n", err)

		return
	}

	for _, d := range documents {
		a.registerIndicatorDocument(d)
	}
}

func (a Agent) registerIndicatorDocument(indicatorsDocument document) {
	registry := fmt.Sprintf(a.RegistryURI+"/v1/register")

	body := bytes.NewBuffer(indicatorsDocument)

	resp, err := a.Client.Post(registry, "text/plain", body)

	if err != nil {
		registrationCount.WithLabelValues("err").Inc()
		log.Printf("could not make http request: %s\n", err)
	} else {
		registrationCount.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()
		if resp.StatusCode != http.StatusOK {
			logErrorResponse(resp)
			return
		}

		closeBodyAndReuseConnection(resp)
	}
}

func logErrorResponse(resp *http.Response) {
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("could not read response body on status %s: %s\n", resp.Status, err)
		return
	}

	log.Printf("received error response from registry: %s\n", string(body))
}

func closeBodyAndReuseConnection(resp *http.Response) {
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

type DocumentFinder struct {
	Glob string
}

func (df DocumentFinder) FindAll() ([]document, error) {
	documentPaths, err := filepath.Glob(df.Glob)
	if err != nil {
		return nil, fmt.Errorf("could not read glob indicator documents: %s/n", err)
	}

	documents := make([]document, 0)
	for _, path := range documentPaths {
		document, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not read indicator document: %s/n", err)
		}

		documents = append(documents, document)
	}

	return documents, nil
}
