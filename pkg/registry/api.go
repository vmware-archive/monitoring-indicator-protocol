package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func NewRegisterHandler(store *DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		defer r.Body.Close()
		documentBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeErrors(w, http.StatusBadRequest, err)
			return
		}

		doc, errs := indicator.ProcessDocument(store.AllPatches(), documentBytes)
		if errs != nil {
			writeErrors(w, http.StatusBadRequest, errs...)
			return
		}

		store.UpsertDocument(doc)

		w.WriteHeader(http.StatusOK)
	}
}

func NewIndicatorDocumentsHandler(store *DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		bytes, _ := marshal(store.AllDocuments())
		fmt.Fprintf(w, string(bytes))
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

func marshal(docs []indicator.Document) ([]byte, error) {
	data := make([]APIV0Document, 0)
	for _, doc := range docs {
		data = append(data, ToAPIV0Document(doc))
	}

	return json.Marshal(data)
}
