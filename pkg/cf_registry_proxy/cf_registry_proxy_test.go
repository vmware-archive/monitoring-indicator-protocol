package cf_registry_proxy_test

import (
	"encoding/json"
	"github.com/cloudfoundry-community/go-cfclient"
	. "github.com/onsi/gomega"
	. "github.com/pivotal/monitoring-indicator-protocol/pkg/cf_registry_proxy"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestIndicatorDocumentsHandler(t *testing.T) {
	g := NewGomegaWithT(t)

	capiClient := &fakeCapiClient{}
	makeFakeCapiClient := func(authToken string)(CapiClient) {
		capiClient.token = authToken
		return capiClient
	}

	t.Run("it sets the token and returns indicator documents", func(t *testing.T) {

		indicatorDocuments := []registry.APIV0Document{{
			APIVersion: "",
			Product:    registry.APIV0Product{},
			Metadata:   map[string]string{"service_instance_guid":"abc-123"},
			Indicators: nil,
			Layout:     registry.APIV0Layout{},
		}}
		handlerFunc := IndicatorDocumentsHandler(&fakeRegistryClient{indicatorDocuments:indicatorDocuments}, makeFakeCapiClient)
		recorder := httptest.NewRecorder()

		request := httptest.NewRequest("GET", "/indicator-documents", nil)
		request.Header.Add("Authorization", "Bearer my-token")
		queries := request.URL.Query()
		queries.Add("service_instance_guid", "abc-123")
		request.URL.RawQuery = queries.Encode()

		handlerFunc(recorder, request)

		g.Expect(capiClient.token).To(Equal("Bearer my-token"))

		body, err := ioutil.ReadAll(recorder.Body)
		g.Expect(err).NotTo(HaveOccurred())

		actualDocuments := make([]registry.APIV0Document, 0)
		err = json.Unmarshal(body, &actualDocuments)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(actualDocuments).To(HaveLen(1))
	})

	t.Run("it filters indicator documents by service_instance_guid from the query param", func(t *testing.T) {
		indicatorDocuments := []registry.APIV0Document{{
			APIVersion: "",
			Product:    registry.APIV0Product{},
			Metadata:   map[string]string{"service_instance_guid":"abc-123"},
			Indicators: nil,
			Layout:     registry.APIV0Layout{},
		},
			{
				APIVersion: "",
				Product:    registry.APIV0Product{},
				Metadata:   map[string]string{"service_instance_guid":"def-456"},
				Indicators: nil,
				Layout:     registry.APIV0Layout{},
			},
		}
		handlerFunc := IndicatorDocumentsHandler(&fakeRegistryClient{indicatorDocuments:indicatorDocuments}, makeFakeCapiClient)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/indicator-documents", nil)
		request.Header.Add("Authorization", "Bearer my-token")
		queries := request.URL.Query()
		queries.Add("service_instance_guid", "def-456")
		request.URL.RawQuery = queries.Encode()

		handlerFunc(recorder, request)

		body, _ := ioutil.ReadAll(recorder.Body)

		actualDocuments := make([]registry.APIV0Document, 0)
		err := json.Unmarshal(body, &actualDocuments)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(actualDocuments).To(HaveLen(1))
		g.Expect(actualDocuments[0]).To(Equal(indicatorDocuments[1]))
	})
}

type fakeRegistryClient struct {
	indicatorDocuments []registry.APIV0Document
}

func (rc fakeRegistryClient) IndicatorDocuments() ([]registry.APIV0Document, error) {
	return rc.indicatorDocuments, nil
}

type fakeCapiClient struct {
	token string
}

func (fakeCapiClient) ServiceInstanceByGuid(string) (cfclient.ServiceInstance, error) {
	return cfclient.ServiceInstance{}, nil
}