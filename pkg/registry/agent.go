package registry

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

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
		log.Print("could not find documents")

		return
	}

	for _, d := range documents {
		a.registerIndicatorDocument(d)
	}
}

func (a Agent) registerIndicatorDocument(indicatorsDocument document) {
	registry := fmt.Sprintf(a.RegistryURI + "/v1alpha1/register")

	body := bytes.NewBuffer(indicatorsDocument)

	// TODO: move to the registry API file
	resp, err := a.Client.Post(registry, "text/plain", body)

	if err != nil {
		registrationCount.WithLabelValues("err").Inc()
		log.Print("could not post to the registry")
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
	_, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("could not read response body on status %s", resp.Status)
		return
	}

	log.Printf("received error response from registry with status: %s", resp.Status)
}

func closeBodyAndReuseConnection(resp *http.Response) {
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

type DocumentFinder struct {
	Glob string
}

func (df DocumentFinder) FindAll() ([]document, error) {
	documentPaths, err := filepath.Glob(df.Glob)
	if err != nil {
		return nil, errors.New("could not read glob indicator documents")
	}

	documents := make([]document, 0)
	for _, path := range documentPaths {
		document, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.New("could not read indicator document")
		}

		documents = append(documents, document)
	}

	return documents, nil
}
