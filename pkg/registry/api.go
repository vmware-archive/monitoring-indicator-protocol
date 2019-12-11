package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"

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

func NewIndicatorDocumentsHandler(store *DocumentStore, statusStore *status_store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		documents := store.FilteredDocuments(r.URL.Query())

		returnedDocuments := make([]APIDocumentResponse, len(documents))
		for i, doc := range documents {
			statusStore.FillStatuses(&doc)
			returnedDocuments[i] = ToAPIDocumentResponse(doc)
		}
		bytes, err := json.Marshal(returnedDocuments)
		if err != nil {
			writeErrors(w, http.StatusInternalServerError, err)
		}
		_, err = fmt.Fprint(w, string(bytes))

		if err != nil {
			log.Printf("error writing to `/indicator-documents`")
		}
	}
}

func writeErrors(w http.ResponseWriter, statusCode int, errors ...error) {
	errorStrings := make([]string, 0)
	for _, e := range errors {
		errorStrings = append(errorStrings, e.Error())
	}

	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorResponse{Errors: errorStrings})
}

type ApiV1UpdateIndicatorStatus struct {
	Name   string  `json:"name"`
	Status *string `json:"status"`
}

func NewIndicatorStatusBulkUpdateHandler(store *status_store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var indicatorStatuses []ApiV1UpdateIndicatorStatus
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bytes, &indicatorStatuses)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		documentID := mux.Vars(r)["documentID"]
		for _, indicatorStatus := range indicatorStatuses {
			store.UpdateStatus(status_store.UpdateRequest{
				Status:        indicatorStatus.Status,
				IndicatorName: indicatorStatus.Name,
				DocumentUID:   documentID,
			})
		}
	}
}

type errorResponse struct {
	Errors []string `json:"errors"`
}
