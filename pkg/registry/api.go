package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		bytes, _ := marshal(store.AllDocuments(), statusStore)
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

type APIV0UpdateIndicatorStatus struct {
	Name   string `json:"name"`
	Status *string `json:"status"`
}

func NewIndicatorStatusBulkUpdateHandler(store *status_store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var indicatorStatuses []APIV0UpdateIndicatorStatus
		err := json.NewDecoder(r.Body).Decode(&indicatorStatuses)
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

func marshal(docs []indicator.Document, statusStore *status_store.Store) ([]byte, error) {
	data := make([]APIV0Document, 0)
	for _, doc := range docs {
		data = append(data, ToAPIV0Document(doc, statusGetterForDoc(statusStore, doc)))
	}

	return json.Marshal(data)
}

func statusGetterForDoc(statusStore *status_store.Store, doc indicator.Document) func(name string) *APIV0IndicatorStatus {
	return func(name string) *APIV0IndicatorStatus {
		status, err := statusStore.StatusFor(doc.UID(), name)
		if err != nil {
			return nil
		}

		return &APIV0IndicatorStatus{
			Value:     status.Status,
			UpdatedAt: status.UpdatedAt,
		}
	}
}
