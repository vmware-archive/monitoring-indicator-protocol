package prometheus_oauth_client

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

type OAuthClientConfig struct {
	Insecure          bool
	OAuthServer       string
	OAuthClientID     string
	OAuthClientSecret string
	Timeout           time.Duration
	Clock             func() time.Time
}

func NewTokenFetcher(config OAuthClientConfig) *oauthTokenFetcher {
	if config.Clock == nil {
		config.Clock = time.Now
	}

	return &oauthTokenFetcher{
		oauthHost:         config.OAuthServer,
		oauthClientID:     config.OAuthClientID,
		oauthClientSecret: config.OAuthClientSecret,
		insecure:          config.Insecure,
		timeout:           config.Timeout,
		getTime:           config.Clock,
	}
}

type oauthTokenFetcher struct {
	oauthHost         string
	oauthClientID     string
	oauthClientSecret string
	insecure          bool

	tokenLock  sync.Mutex
	token      string
	expiration time.Time
	timeout    time.Duration
	getTime    func() time.Time
}

func (c *oauthTokenFetcher) GetClientToken() (string, error) {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	if c.tokenIsValid() {
		return c.token, nil
	}

	v := make(url.Values)
	v.Set("client_id", c.oauthClientID)
	v.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(
		"POST",
		c.oauthHost,
		strings.NewReader(v.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.URL.Path = "/oauth/token"

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.URL.User = url.UserPassword(c.oauthClientID, c.oauthClientSecret)

	return c.doTokenRequest(req)
}

func (c *oauthTokenFetcher) doTokenRequest(req *http.Request) (string, error) {
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
	c.expiration = c.getTime().Add(c.timeout)
	return c.token, nil
}

func (c *oauthTokenFetcher) tokenIsValid() bool {
	return c.token != "" && c.expiration.After(c.getTime())
}
