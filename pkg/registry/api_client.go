package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type APIClient interface {
	IndicatorDocuments() ([]APIV0Document, error)
}

type apiClient struct {
	serverURL string
	client    *http.Client
}

func NewAPIClient(serverURL string, client *http.Client) APIClient {
	return &apiClient{
		serverURL: serverURL,
		client:    client,
	}
}

func (c *apiClient) indicatorResponse() ([]byte, error) {
	resp, err := c.client.Get(c.serverURL + "/v1/indicator-documents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *apiClient) IndicatorDocuments() ([]APIV0Document, error) {
	payload, e := c.indicatorResponse()
	if e != nil {
		return nil, fmt.Errorf("failed to get indicator documents: %s\n", e)
	}

	var d []APIV0Document
	err := json.Unmarshal(payload, &d)
	return d, err
}
