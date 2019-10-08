package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ghodss/yaml"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
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
		return nil, errors.New(fmt.Sprintf("failed to get indicator documents: %s", e))
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

	if resp.StatusCode != 200 {
		return fmt.Errorf("error response from status updates: %s", resp.Status)
	}

	return nil
}

func (c *RegistryApiClient) Register(document v1.IndicatorDocument) error {
	docBytes, err := yaml.Marshal(document)
	if err != nil {
		return err
	}
	resp, err := c.client.Post(c.serverURL+"/v1/register", "application/yml", bytes.NewReader(docBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
