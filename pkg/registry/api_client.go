package registry

import (
	"encoding/json"
	"fmt"
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

func (c *APIClient) indicatorResponse() ([]byte, error) {
	resp, err := c.client.Get(c.serverURL + "/v1/indicator-documents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *APIClient) IndicatorDocuments() ([]APIV0Document, error) {
	payload, e := c.indicatorResponse()
	if e != nil {
		return nil, fmt.Errorf("failed to get indicator documents: %s\n", e)
	}

	println(string(payload))

	var d []APIV0Document
	err := json.Unmarshal(payload, &d)
	return d, err
}
