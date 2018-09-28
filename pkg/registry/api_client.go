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

func (c *APIClient) IndicatorDocuments() ([]APIV0Document, error) {
	resp, err := c.client.Get(c.serverURL + "/v1/indicator-documents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s\n", err)
	}

	return unmarshalDocuments(body)
}

func unmarshalDocuments(payload []byte) ([]APIV0Document, error) {
	var d []APIV0Document
	err := json.Unmarshal(payload, &d)
	return d, err
}
