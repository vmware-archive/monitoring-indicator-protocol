package main

import (
	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/prometheus_uaa_client"
	"code.cloudfoundry.org/indicators/pkg/verification"
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	stdOut := log.New(os.Stdout, "", 0)
	stdErr := log.New(os.Stderr, "", 0)

	flagSet := flag.NewFlagSet("validator", flag.ErrorHandling(0))
	indicatorsFile := flagSet.String("indicators", "", "file path of indicators yml (see https://github.com/cloudfoundry-incubator/indicators)")
	logCacheURL := flagSet.String("log-cache-url", "", "the log-cache url (e.g. https://log-cache.system.cfapp.com")
	deployment := flagSet.String("deployment", "", "the source deployment of metrics emitted to loggregator. replaces any $deployment metadata")
	uaaURL := flagSet.String("uaa-url", "", "UAA server host (e.g. https://uaa.my-pcf.com)")
	uaaClient := flagSet.String("log-cache-client", "", "the UAA client which has access to log-cache (doppler.firehose or logs.admin scope")
	uaaClientSecret := flagSet.String("log-cache-client-secret", "", "the client secret")
	insecure := flagSet.Bool("k", false, "skips ssl verification (insecure)")

	flagSet.Parse(os.Args[1:])

	document, err := indicator.ReadFile(*indicatorsFile, indicator.OverrideMetadata(map[string]string{"deployment": *deployment}))
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}

	tokenFetcher := prometheus_uaa_client.NewUAATokenFetcher(prometheus_uaa_client.UAAClientConfig{
		*insecure,
		*uaaURL,
		*uaaClient,
		*uaaClientSecret,
		time.Minute,
	})

	prometheusClient, err := prometheus_uaa_client.Build(*logCacheURL, tokenFetcher.GetClientToken, *insecure)
	if err != nil {
		log.Fatalf("could not create prometheus client: %s\n", err)
	}

	stdOut.Println("---------------------------------------------------------------------------------------------")
	stdOut.Printf("Querying current value for %d indicators in log-cache \n", len(document.Indicators))
	stdOut.Println("---------------------------------------------------------------------------------------------")

	failedIndicators := make([]indicator.Indicator, 0)

	for _, ind := range document.Indicators {
		stdOut.Println()

		stdOut.Printf("Querying log-cache for indicator with name \"%s\"", ind.Name)
		stdOut.Printf("  query: %s", ind.PromQL)

		result, err := verification.VerifyIndicator(ind, prometheusClient)
		if err != nil {
			stdErr.Println("  " + err.Error())
			failedIndicators = append(failedIndicators, ind)
			continue
		}

		for _, s := range result.Series {
			stdOut.Printf("  [%s] -> [%s] \n", s.Labels, s.Points)
		}

		if result.MaxNumberOfPoints == 0 {
			stdErr.Println("  no data points found")
			failedIndicators = append(failedIndicators, ind)
		}
	}

	if len(failedIndicators) > 0 {
		separator(stdOut)
		stdErr.Println("---------------------------------------------------------------------------------------------")
		stdErr.Println("VALIDATION FAILURE")
		stdErr.Printf("  Could not find %d indicators in log-cache \n", len(failedIndicators))
		stdErr.Println("  Both operators and platform observability tools such as PCF Healthwatch rely on the")
		stdErr.Println("  existence of this data. Perhaps a metric name changed, or refactored code")
		stdErr.Println("  is failing to emit.")
		stdErr.Println("---------------------------------------------------------------------------------------------")
		separator(stdOut)

		os.Exit(1)
	}

	separator(stdOut)
	stdOut.Println("---------------------------------------------------------------------------------------------")
	stdOut.Println("  All indicator data found")
	stdOut.Println("---------------------------------------------------------------------------------------------")
	separator(stdOut)
}

func separator(stdOut *log.Logger) {
	stdOut.Println()
	stdOut.Println()
}
