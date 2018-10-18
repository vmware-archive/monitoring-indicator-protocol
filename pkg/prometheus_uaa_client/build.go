package prometheus_uaa_client

import (
	"context"
	"crypto/tls"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"net/http"
	"net/url"
	"time"
)

type wrappedClient struct {
	tf *uaaTokenFetcher
	prometheus api.Client
}

func (c wrappedClient) URL(ep string, args map[string]string) *url.URL {
	return c.prometheus.URL(ep, args)
}

func (c wrappedClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	token, err := c.tf.GetClientToken()
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Authorization", token)

	return c.prometheus.Do(ctx, req)
}

func Build(url string, uaaHost string, uaaClientID string, uaaClientSecret string, insecure bool) (v1.API, error) {
	prometheusClient, err := api.NewClient(api.Config{
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

	config := UAAClientConfig{
		insecure,
		uaaHost,
		uaaClientID,
		uaaClientSecret,
		time.Minute,
	}

	tokenFetcher := NewUAATokenFetcher(config)

	if err != nil {
		return nil, err
	}

	return v1.NewAPI(wrappedClient{tokenFetcher, prometheusClient}), err
}
