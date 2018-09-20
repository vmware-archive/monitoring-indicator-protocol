package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/verification"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

func main() {
	stdOut := log.New(os.Stdout, "", 0)
	stdErr := log.New(os.Stderr, "", 0)

	flagSet := flag.NewFlagSet("validator", flag.ErrorHandling(0))
	indicatorsFile := flagSet.String("indicators", "", "file path of indicators yml (see https://github.com/cloudfoundry-incubator/indicators)")
	logCacheURL := flagSet.String("log-cache-url", "", "the log-cache url (e.g. https://log-cache.system.cfapp.com")
	deployment := flagSet.String("deployment", "", "the source deployment of metrics emitted to loggregator")
	uaaURL := flagSet.String("uaa-url", "", "UAA server host (e.g. https://uaa.my-pcf.com)")
	uaaClient := flagSet.String("log-cache-client", "", "the UAA client which has access to log-cache (doppler.firehose or logs.admin scope")
	uaaClientSecret := flagSet.String("log-cache-client-secret", "", "the client secret")
	lookback := flagSet.String("lookback", "1m", "the promQL query time period")
	insecure := flagSet.Bool("k", false, "skips ssl verification (insecure)")

	flagSet.Parse(os.Args[1:])

	document, err := indicator.ReadFile(*indicatorsFile)
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}

	prometheusClient, err := buildPrometheusClient(*logCacheURL, *uaaURL, *uaaClient, *uaaClientSecret, *insecure)
	if err != nil {
		log.Fatalf("could not create prometheus client: %s\n", err)
	}

	stdOut.Println("---------------------------------------------------------------------------------------------")
	stdOut.Printf("Checking for existence of %d metrics in log-cache \n", len(document.Metrics))
	stdOut.Println("---------------------------------------------------------------------------------------------")

	failedMetrics := make([]indicator.Metric, 0)

	for _, m := range document.Metrics {
		stdOut.Println()

		query := verification.FormatQuery(m, *deployment, *lookback)

		stdOut.Printf("Querying log-cache for metric with name \"%s\" and source_id \"%s\"", m.Name, m.SourceID)
		stdOut.Printf("  query: %s", query)

		result, err := verification.VerifyMetric(m, query, prometheusClient)
		if err != nil {
			stdErr.Println("  " + err.Error())
			failedMetrics = append(failedMetrics, m)
			continue
		}

		for _, s := range result.Series {
			stdOut.Printf("  [%s] -> [%s] \n", s.Labels, s.Points)
		}

		if result.MaxNumberOfPoints == 0 {
			stdErr.Println("  no data points found")
			failedMetrics = append(failedMetrics, m)
		}
	}

	if len(failedMetrics) > 0 {
		separator(stdOut)
		stdErr.Println("---------------------------------------------------------------------------------------------")
		stdErr.Println("VALIDATION FAILURE")
		stdErr.Printf("  Could not find %d metrics in log-cache \n", len(failedMetrics))
		stdErr.Println("  Both operators and platform observability tools such as PCF Healthwatch rely on the")
		stdErr.Println("  existence of these metrics. Perhaps the name of the metric changed, or refactored code")
		stdErr.Println("  is failing to emit.")
		stdErr.Println("---------------------------------------------------------------------------------------------")
		separator(stdOut)

		os.Exit(1)
	}

	separator(stdOut)
	stdOut.Println("---------------------------------------------------------------------------------------------")
	stdOut.Println("  All metrics found")
	stdOut.Println("---------------------------------------------------------------------------------------------")
	separator(stdOut)
}

func separator(stdOut *log.Logger) {
	stdOut.Println()
	stdOut.Println()
}

func buildPrometheusClient(logCacheURL string, uaaHost string, uaaClientID string, uaaClientSecret string, insecure bool) (v1.API, error) {

	client, err := api.NewClient(api.Config{
		Address: logCacheURL,
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

	return v1.NewAPI(c), err
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
