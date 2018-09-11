package registry

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"
	"code.cloudfoundry.org/indicators/pkg/indicator"
)


func NewRegisterHandler(d *DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		defer r.Body.Close()
		documentBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeErrors(w, http.StatusBadRequest, err)
			return
		}

		doc, err := indicator.ReadIndicatorDocument(documentBytes)
		if err != nil {
			writeErrors(w, http.StatusBadRequest, err)
			return
		}

		if doc.Labels == nil {
			doc.Labels = make(map[string]string)
		}

		labelValues := r.URL.Query()
		for k, v := range labelValues {
			doc.Labels[k] = v[0]

			if len(v) > 1 {
				writeErrors(w, http.StatusBadRequest, fmt.Errorf("label %s has too many values", k))
				return
			}
		}

		if doc.Labels["deployment"] == "" {
			writeErrors(w, http.StatusBadRequest, fmt.Errorf("deployment query parameter is required"))
			return
		}

		errs := indicator.Validate(doc)
		if len(errs) > 0 {
			writeErrors(w, http.StatusUnprocessableEntity, errs...)
			return
		}

		d.Upsert(doc.Labels, doc.Indicators)

		w.WriteHeader(http.StatusOK)
	}
}

func writeErrors(w http.ResponseWriter, statusCode int, errors ...error) {
	errorStrings := make([]string, 0)
	for _, e := range errors {
		errorStrings = append(errorStrings, e.Error())
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse{Errors: errorStrings})
}

type errorResponse struct {
	Errors []string `json:"errors"`
}

func marshal(docs []Document) ([]byte, error) {
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

func NewIndicatorDocumentsHandler(d *DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		bytes, _ := marshal(d.All())
		fmt.Fprintf(w, string(bytes))
	}
}
