package prometheus_oauth_client

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_client"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type wrappedClient struct {
	fetchToken TokenFetcherFunc
	prometheus api.Client
}

func (c wrappedClient) URL(ep string, args map[string]string) *url.URL {
	return c.prometheus.URL(ep, args)
}

func (c wrappedClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	token, err := c.fetchToken()
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Authorization", token)

	return c.prometheus.Do(ctx, req)
}

type TokenFetcherFunc func() (string, error)

func Build(url string, fetchToken TokenFetcherFunc, insecure bool) (*prometheus_client.PrometheusClient, error) {
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

	prometheusAPI := v1.NewAPI(wrappedClient{fetchToken, prometheusClient})
	return &prometheus_client.PrometheusClient{
		Api: prometheusAPI,
	}, err
}
