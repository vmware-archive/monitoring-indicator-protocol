package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type RegistryApiClient struct {
	serverURL string
	client    *http.Client
}

func NewAPIClient(serverURL string, client *http.Client) *RegistryApiClient {
	return &RegistryApiClient{
		serverURL: serverURL,
		client:    client,
	}
}

func (c *RegistryApiClient) IndicatorDocuments() ([]APIDocumentResponse, error) {
	payload, e := c.indicatorResponse()
	if e != nil {
		return nil, errors.New("failed to get indicator documents")
	}

	var d []APIDocumentResponse
	err := json.Unmarshal(payload, &d)

	return d, err
}

func (c *RegistryApiClient) indicatorResponse() ([]byte, error) {
	resp, err := c.client.Get(c.serverURL + "/v1/indicator-documents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *RegistryApiClient) BulkStatusUpdate(statusUpdates []APIV0UpdateIndicatorStatus, documentId string) error {
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
		return errors.New("error sending status updates")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error response from status updates: %s", resp.Status)
	}

	return nil
}

func (c *RegistryApiClient) AddIndicatorDocument(document []byte) error {
	body := bytes.NewBuffer(document)
	resp, err := c.client.Post(c.serverURL + "/v1/register", "text/plain", body)
	if err != nil {
		return err
	}

	registrationCount.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-successful response from registry: %d", resp.StatusCode)
	}

	closeBodyAndReuseConnection(resp)
	return nil
}
