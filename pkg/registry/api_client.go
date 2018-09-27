package registry

import (
	"io/ioutil"
	"net/http"
)

type APIClient struct {
	serverURL string
	client    *http.Client
}

func NewAPIClient(serverURL string, client *http.Client) *APIClient {
	return &APIClient{
		serverURL: serverURL,
		client:    client,
	}
}

type IndicatorDocuments []byte

func (c *APIClient) IndicatorDocuments() (IndicatorDocuments, error) {
	resp, err := c.client.Get(c.serverURL + "/v1/indicator-documents")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
