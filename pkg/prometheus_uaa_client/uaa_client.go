package prometheus_uaa_client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type UAAClientConfig struct {
	Insecure        bool
	UAAHost         string
	UAAClientID     string
	UAAClientSecret string
	Timeout         time.Duration
}

func NewUAATokenFetcher(config UAAClientConfig) *uaaTokenFetcher {
	return &uaaTokenFetcher{
		uaaHost:         config.UAAHost,
		uaaClientID:     config.UAAClientID,
		uaaClientSecret: config.UAAClientSecret,
		insecure:        config.Insecure,
		timeout:         config.Timeout,
	}
}

type uaaTokenFetcher struct {
	uaaHost         string
	uaaClientID     string
	uaaClientSecret string
	insecure        bool

	tokenLock  sync.Mutex
	token      string
	expiration time.Time
	timeout    time.Duration
}

func (c *uaaTokenFetcher) GetClientToken() (string, error) {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	if c.tokenIsValid() {
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

func (c *uaaTokenFetcher) doTokenRequest(req *http.Request) (string, error) {
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
	c.expiration = time.Now().Add(c.timeout)
	return c.token, nil
}

func (c *uaaTokenFetcher) tokenIsValid() bool {
	return c.token != "" && c.expiration.After(time.Now())
}
