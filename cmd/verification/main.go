package main

import (
	"flag"
	"log"
	"os"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_oauth_client"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/verification"
)

func main() {
	stdOut := log.New(os.Stdout, "", 0)
	stdErr := log.New(os.Stderr, "", 0)
	l := log.New(os.Stderr, "", 0)

	indicatorsFilePath := flag.String("indicators", "", "file path of indicators yml (see https://github.com/cloudfoundry-incubator/indicators)")
	metadata := flag.String("metadata", "", "metadata to override (e.g. --metadata deployment=my-test-deployment,source_id=metric-forwarder)")
	prometheusURI := flag.String("query-endpoint", "", "the query url of a Prometheus compliant store (e.g. https://metric-store.system.cfapp.com")
	authorization := flag.String("authorization", "", "the authorization header sent to prometheus (e.g. 'bearer abc-123')")
	insecure := flag.Bool("k", false, "skips ssl verification (insecure)")
	flag.Parse()

	checkRequiredFlagsArePresent(*indicatorsFilePath, *prometheusURI, *authorization)

	document, err := indicator.ReadFile(*indicatorsFilePath, indicator.OverrideMetadata(indicator.ParseMetadata(*metadata)))
	if err != nil {
		l.Fatal(err)
	}

	tokenFetcher := func() (string, error) { return *authorization, nil }

	prometheusClient, err := prometheus_oauth_client.Build(*prometheusURI, tokenFetcher, *insecure)
	if err != nil {
		l.Fatalf("could not create prometheus client: %s\n", err)
	}

	printHorizontalLine(stdOut)
	stdOut.Printf("Querying current value for %d indicators in Prometheus compliant store \n", len(document.Spec.Indicators))
	printHorizontalLine(stdOut)

	failedIndicators := make([]v1.IndicatorSpec, 0)

	for _, ind := range document.Spec.Indicators {
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
		printHorizontalLine(stdErr)
		stdErr.Println("VALIDATION FAILURE")
		stdErr.Printf("  Could not find %d indicators in %s \n", len(failedIndicators), *prometheusURI)
		stdErr.Println("  Both operators and platform observability tools such as PCF Healthwatch rely on the")
		stdErr.Println("  existence of this data. Perhaps a metric name changed, or refactored code")
		stdErr.Println("  is failing to emit.")
		printHorizontalLine(stdErr)
		separator(stdOut)

		os.Exit(1)
	}

	separator(stdOut)
	printHorizontalLine(stdOut)
	stdOut.Println("  All indicator data found")
	printHorizontalLine(stdOut)
	separator(stdOut)
}

func checkRequiredFlagsArePresent(indicatorsFilePath string, prometheusURI string, prometheusAuthorization string) {
	stdErr := log.New(os.Stderr, "", 0)

	exitOnEmpty := func(v string, name string) {
		if v == "" {
			stdErr.Printf("%s is required\n\n", name)
			stdErr.Printf("Usage:\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	exitOnEmpty(indicatorsFilePath, "indicators")
	exitOnEmpty(prometheusURI, "query-endpoint")
	exitOnEmpty(prometheusAuthorization, "authorization")
}

func separator(logger *log.Logger) {
	logger.Println()
	logger.Println()
}

func printHorizontalLine(logger *log.Logger) {
	logger.Println("---------------------------------------------------------------------------------------------")
}
