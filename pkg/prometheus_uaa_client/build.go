package prometheus_uaa_client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"net/http"
	"net/url"
	"strings"
)

func Build(url string, uaaHost string, uaaClientID string, uaaClientSecret string, insecure bool) (v1.API, error) {
	c, err := NewUaaClient(url, insecure, uaaHost, uaaClientID, uaaClientSecret)

	if err != nil {
		return nil, err
	}

	return v1.NewAPI(c), err
}

func NewUaaClient(url string, insecure bool, uaaHost string, uaaClientID string, uaaClientSecret string) (*uaaClient, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
		RoundTripper: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	})

	if err != nil {
		return nil, err
	}
	c := &uaaClient{
		Client:          client,
		uaaHost:         uaaHost,
		uaaClientID:     uaaClientID,
		uaaClientSecret: uaaClientSecret,
		insecure:        insecure,
	}
	return c, nil
}

type uaaClient struct {
	api.Client
	uaaHost         string
	uaaClientID     string
	uaaClientSecret string
	insecure        bool
	token           string
}

func (c *uaaClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	token, err := c.getClientToken()
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Authorization", token)

	return c.Client.Do(ctx, req)
}

func (c *uaaClient) getClientToken() (string, error) {
	if c.token != "" {
		return c.token, nil
	}

	v := make(url.Values)
	v.Set("client_id", c.uaaClientID)
	v.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(
		"POST",
		c.uaaHost,
		strings.NewReader(v.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.URL.Path = "/oauth/token"

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.URL.User = url.UserPassword(c.uaaClientID, c.uaaClientSecret)

	return c.doTokenRequest(req)
}

func (c *uaaClient) doTokenRequest(req *http.Request) (string, error) {
	client := http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.insecure,
		},
	}}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code from Oauth2 server %d", resp.StatusCode)
	}

	token := struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", fmt.Errorf("failed to unmarshal response from Oauth2 server: %s", err)
	}

	c.token = fmt.Sprintf("%s %s", token.TokenType, token.AccessToken)
	return c.token, nil
}
