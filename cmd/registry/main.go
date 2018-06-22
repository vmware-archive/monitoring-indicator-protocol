package main

import (
	"flag"
	"net/http"
	"fmt"

	"github.com/gorilla/mux"
	"code.cloudfoundry.org/cf-indicators/pkg/registry"
	"io/ioutil"
	"log"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"encoding/json"
)

func main() {
	port := flag.Int("port", -1, "Port to expose regisration endpoints")
	flag.Parse()

	documentStore := registry.NewDocumentStore()

	r := mux.NewRouter()
	r.HandleFunc("/v1/register", NewRegisterHandler(documentStore))
	r.HandleFunc("/v1/indicator-documents", NewIndicatorDocumentsHandler(documentStore))

	http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
}

func NewRegisterHandler(d *registry.DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		labelValues := r.URL.Query()
		labels := make(map[string]string)
		for k, v := range labelValues {
			labels[k] = v[0]

			if len(v) > 1 {
				log.Fatal("//TODO test me")
			}
		}

		defer r.Body.Close()
		documentBytes, _ := ioutil.ReadAll(r.Body)
		//TODO: errors

		doc, _ := indicator.ReadIndicatorDocument(documentBytes)
		//TODO: errs := indicator.Validate(doc)

		d.Insert(labels, doc.Indicators)

		w.WriteHeader(http.StatusOK)
	}
}

func marshal(docs []registry.Document) ([]byte, error) {
	type metric struct {
		Title       string `json:"title"`
		Origin      string `json:"origin"`
		SourceID    string `json:"source_id"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Frequency   string `json:"frequency"`
	}
	type threshold struct {
		Level    string  `json:"level"`
		Dynamic  bool    `json:"dynamic"`
		Operator string  `json:"operator"`
		Value    float64 `json:"value"`
	}
	type indicator struct {
		Name        string      `json:"name"`
		Title       string      `json:"title"`
		Description string      `json:"description"`
		PromQL      string      `json:"promql"`
		Thresholds  []threshold `json:"thresholds"`
		Metrics     []metric    `json:"metrics"`
		Response    string      `json:"response"`
		Measurement string      `json:"measurement"`
	}
	type document struct {
		Labels     map[string]string `json:"labels"`
		Indicators []indicator       `json:"indicators"`
	}

	data := make([]document, 0)
	for _, doc := range docs {
		indicators := make([]indicator, 0)
		for _, i := range doc.Indicators {
			thresholds := make([]threshold, 0)
			for _, t := range i.Thresholds {
				thresholds = append(thresholds, threshold{
					Level:    t.Level,
					Dynamic:  t.Dynamic,
					Operator: t.Operator.String(),
					Value:    t.Value,
				})
			}

			metrics := make([]metric, 0)
			for _, m := range i.Metrics {
				metrics = append(metrics, metric{
					Title:       m.Title,
					Origin:      m.Origin,
					SourceID:    m.SourceID,
					Name:        m.Name,
					Type:        m.Type,
					Description: m.Description,
					Frequency:   m.Frequency,
				})
			}

			indicators = append(indicators, indicator{
				Name:        i.Name,
				Title:       i.Title,
				Description: i.Description,
				PromQL:      i.PromQL,
				Thresholds:  thresholds,
				Metrics:     metrics,
				Response:    i.Response,
				Measurement: i.Measurement,
			})
		}

		data = append(data, document{
			Labels:     doc.Labels,
			Indicators: indicators,
		})
	}

	return json.Marshal(data)
}

func NewIndicatorDocumentsHandler(d *registry.DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		bytes, _ := marshal(d.All())
		fmt.Fprintf(w, string(bytes))
	}
}
