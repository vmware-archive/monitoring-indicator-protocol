package registry

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
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

	apiClient := NewAPIClient(a.RegistryURI, a.Client)

	for _, d := range documents {
		err := apiClient.AddIndicatorDocument(d)
		if err != nil {
			registrationCount.WithLabelValues("err").Inc()
			log.Print(err)
		}
	}
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
