package cf_registry_proxy

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

type CapiClient interface {
	ServiceInstanceByGuid(string) (cfclient.ServiceInstance, error)
}

type RegistryClient interface {
	IndicatorDocuments() ([]registry.APIV0Document, error)
}

func IndicatorDocumentsHandler(registryClient RegistryClient, makeCapiClient func(string) CapiClient) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		capiClient := makeCapiClient(request.Header.Get("Authorization"))

		requestedServiceInstanceGuid := request.URL.Query().Get("service_instance_guid")
		_, err := capiClient.ServiceInstanceByGuid(requestedServiceInstanceGuid)
		if err != nil {
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		documents, _ := registryClient.IndicatorDocuments()
		filteredDocuments := []registry.APIV0Document{}
		for _, document := range documents {
			if document.Metadata["service_instance_guid"] == requestedServiceInstanceGuid {
				filteredDocuments = append(filteredDocuments, document)
			}
		}

		json.NewEncoder(writer).Encode(filteredDocuments)
	}
}
