package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type apiClient struct {
	serverURL string
	client    *http.Client
}

func NewAPIClient(serverURL string, client *http.Client) *apiClient {
	return &apiClient{
		serverURL: serverURL,
		client:    client,
	}
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

func (c *apiClient) indicatorResponse() ([]byte, error) {
	resp, err := c.client.Get(c.serverURL + "/v1/indicator-documents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *apiClient) BulkStatusUpdate(statusUpdates []APIV0UpdateIndicatorStatus, documentId string) error {
	updateBytes, err := json.Marshal(statusUpdates)
	body := bytes.NewBuffer(updateBytes)
	if err != nil {
		return fmt.Errorf("error marshaling status updates: %s", err)
	}
	resp, err := c.client.Post(
		fmt.Sprintf("%s/v1/indicator-documents/%s/bulk_status", c.serverURL, documentId),
		"application/json",
		body,
	)

	if err != nil {
		return fmt.Errorf("error sending status updates: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error response from status updates: %s", resp.Status)
	}

	return nil
}
