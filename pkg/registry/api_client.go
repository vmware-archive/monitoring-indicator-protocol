package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	utils "github.com/pivotal/monitoring-indicator-protocol/pkg"
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

func (c *RegistryApiClient) IndicatorDocuments() ([]APIV0Document, error) {
	payload, e := c.indicatorResponse()
	if e != nil {
		errString := utils.SanitizeUrl(e, c.serverURL, "failed to get indicator documents")
		return nil, fmt.Errorf(errString)
	}

	var d []APIV0Document
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
		errMessage := utils.SanitizeUrl(err, c.serverURL, "Error sending status updates")
		return fmt.Errorf(errMessage)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error response from status updates: %s", resp.Status)
	}

	return nil
}
