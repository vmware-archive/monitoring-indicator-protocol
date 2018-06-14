package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"code.cloudfoundry.org/cf-indicators/pkg/validation"
	"code.cloudfoundry.org/go-log-cache"
)

func main() {
	stdOut := log.New(os.Stdout, "", 0)
	stdErr := log.New(os.Stderr, "", 0)

	flagSet := flag.NewFlagSet("validator", flag.ErrorHandling(0))
	indicatorsFile := flagSet.String("indicators", "", "file path of indicators yml (see https://github.com/cloudfoundry-incubator/cf-indicators)")
	logCacheURL := flagSet.String("log-cache-url", "", "the log-cache url (e.g. https://log-cache.system.cfapp.com")
	deployment := flagSet.String("deployment", "", "the source deployment of metrics emitted to loggregator")
	uaaURL := flagSet.String("uaa-url", "", "UAA server host (e.g. https://uaa.my-pcf.com)")
	uaaClient := flagSet.String("log-cache-client", "", "the UAA client which has access to log-cache (doppler.firehose or logs.admin scope")
	uaaClientSecret := flagSet.String("log-cache-client-secret", "", "the client secret")
	insecure := flagSet.Bool("k", false, "skips ssl verification (insecure)")

	flagSet.Parse(os.Args[1:])

	metrics, err := readMetrics(*indicatorsFile)
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}

	logCache := buildLogCacheClient(*uaaURL, *uaaClient, *uaaClientSecret, *insecure)

	stdOut.Println("---------------------------------------------------------------------------------------------")
	stdOut.Printf("Checking for existence of %d metrics in log-cache \n", len(metrics))
	stdOut.Println("---------------------------------------------------------------------------------------------")

	failedMetrics := make([]indicator.Metric, 0)

	for _, m := range metrics {
		stdOut.Println()

		query := validation.FormatQuery(m, *deployment)

		stdOut.Printf("Querying log-cache for metric with name \"%s\" and source_id \"%s\"", m.Name, m.SourceID)
		stdOut.Printf("  query: %s", query)

		result, err := validation.VerifyMetric(m, query, *logCacheURL+"/v1/promql", logCache)
		if err != nil {
			stdErr.Println("  " + err.Error())
			failedMetrics = append(failedMetrics, m)
			continue
		}

		for _, s := range result.Series {
			stdOut.Printf("  [%s] -> [%s] \n", formatSeriesID(s.Labels), formatSeriesPoints(s.Points))
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

func formatSeriesPoints(values []float64) string {
	stringValues := make([]string, 0)

	for _, v := range values {
		stringValues = append(stringValues, fmt.Sprintf("%+v", v))
	}

	return strings.Join(stringValues, " ")
}

func formatSeriesID(labels map[string]string) string {
	labelParts := make([]string, 0)

	for k, v := range labels {
		labelParts = append(labelParts, k+":"+v)
	}

	return strings.Join(labelParts, " ")
}

func separator(stdOut *log.Logger) {
	stdOut.Println()
	stdOut.Println()
}

func buildLogCacheClient(uaaHost string, uaaClient string, uaaClientSecret string, insecure bool) *logcache.Oauth2HTTPClient {
	return logcache.NewOauth2HTTPClient(uaaHost, uaaClient, uaaClientSecret,
		logcache.WithOauth2HTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecure,
				},
			},
			Timeout: 10 * time.Second,
		}),
	)
}

func readMetrics(indicatorsFile string) ([]indicator.Metric, error) {
	fileBytes, err := ioutil.ReadFile(indicatorsFile)
	if err != nil {
		return nil, err
	}

	indicatorDocument, err := indicator.ReadIndicatorDocument(fileBytes)
	if err != nil {
		return nil, err
	}

	validationErrors := indicator.Validate(indicatorDocument)
	if len(validationErrors) > 0 {

		log.Println("validation for indicator file failed")
		for _, e := range validationErrors {
			log.Printf("- %s \n", e.Error())
		}

		os.Exit(1)
	}

	return indicatorDocument.Metrics, nil
}
