package main

import (
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_uaa_client"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/verification"
	"flag"
	"log"
	"os"
)

func main() {
	stdOut := log.New(os.Stdout, "", 0)
	stdErr := log.New(os.Stderr, "", 0)

	flagSet := flag.NewFlagSet("validator", flag.ErrorHandling(0))
	indicatorsFilePath := flagSet.String("indicators", "", "file path of indicators yml (see https://github.com/cloudfoundry-incubator/indicators)")
	metadata := flagSet.String("metadata", "", "metadata to overide (e.g. --metadata deployment=my-test-deployment,source_id=metric-forwarder)")
	prometheusURL := flagSet.String("query-endpoint", "", "the query url of a Prometheus compliant store (e.g. https://log-cache.system.cfapp.com")
	authorization := flagSet.String("authorization", "", "the authorization header sent to prometheus (e.g. 'bearer abc-123')")
	insecure := flagSet.Bool("k", false, "skips ssl verification (insecure)")
	flagSet.Parse(os.Args[1:])

	document, err := indicator.ReadFile(*indicatorsFilePath, indicator.OverrideMetadata(indicator.ParseMetadata(*metadata)))
	if err != nil {
		log.Fatalf("could not read indicators document: %s\n", err)
	}

	tokenFetcher := func() (string, error) { return *authorization, nil }

	prometheusClient, err := prometheus_uaa_client.Build(*prometheusURL, tokenFetcher, *insecure)
	if err != nil {
		log.Fatalf("could not create prometheus client: %s\n", err)
	}

	stdOut.Println("---------------------------------------------------------------------------------------------")
	stdOut.Printf("Querying current value for %d indicators in Prometheus compliant store \n", len(document.Indicators))
	stdOut.Println("---------------------------------------------------------------------------------------------")

	failedIndicators := make([]indicator.Indicator, 0)

	for _, ind := range document.Indicators {
		stdOut.Println()

		stdOut.Printf("Querying for indicator with name \"%s\"", ind.Name)
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
		stdErr.Printf("  Could not find %d indicators in %s \n", len(failedIndicators), *prometheusURL)
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
