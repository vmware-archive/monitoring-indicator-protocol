package cf_registry_proxy

import (
	"encoding/json"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"net/http"
)

type CapiClient interface {
	ServiceInstanceByGuid(string) (cfclient.ServiceInstance, error)
}

type RegistryClient interface {
	IndicatorDocuments() ([]registry.APIV0Document, error)
}

func IndicatorDocumentsHandler(registryClient RegistryClient, makeCapiClient func(string)(CapiClient)) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		makeCapiClient(request.Header.Get("Authorization"))

		documents, _ := registryClient.IndicatorDocuments()
		requestedServiceInstanceGuid := request.URL.Query().Get("service_instance_guid")

		filteredDocuments := []registry.APIV0Document{}
		for _, document := range documents {
			if document.Metadata["service_instance_guid"] == requestedServiceInstanceGuid {
				filteredDocuments = append(filteredDocuments, document)
			}
		}

		json.NewEncoder(writer).Encode(filteredDocuments)
	}
}

