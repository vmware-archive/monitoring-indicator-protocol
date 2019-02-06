package prometheus_alerts

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/registry"
)

type ControllerConfig struct {
	RegistryAPIClient   registry.APIClient
	PrometheusAPIClient PrometheusClient
	OutputDirectory     string
}

type PrometheusClient interface {
	Reload() error
}

type prometheusClient struct {
	prometheusURI string
	client        *http.Client
}

func (p *prometheusClient) Reload() error {
	buffer := bytes.NewBuffer(nil)
	resp, err := p.client.Post(fmt.Sprintf("%s/-/reload", p.prometheusURI), "", buffer)
	if err != nil  {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("received %v response from prometheus: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func NewPrometheusClient(prometheusURI string, client *http.Client) PrometheusClient {
	return &prometheusClient{
		prometheusURI: prometheusURI,
		client:        client,
	}
}

type Controller interface {
	Update() error
}

func NewController(c ControllerConfig) Controller {
	return controller{c}
}

type controller struct {
	ControllerConfig
}

func (c controller) Update() error {
	documents, err := c.RegistryAPIClient.IndicatorDocuments()
	if err != nil {
		return fmt.Errorf("failed to fetch indicator documents, %s", err)
	}

	writeDocuments(formatDocuments(documents), c.OutputDirectory)
	return c.PrometheusAPIClient.Reload()
}

func formatDocuments(documents []registry.APIV0Document) []indicator.Document {
	formattedDocuments := make([]indicator.Document, 0)
	for _, d := range documents {
		formattedDocuments = append(formattedDocuments, convertDocument(d))
	}

	return formattedDocuments
}

func convertDocument(d registry.APIV0Document) indicator.Document {
	indicators := make([]indicator.Indicator, 0)
	for _, i := range d.Indicators {
		indicators = append(indicators, convertIndicator(i))
	}

	return indicator.Document{
		Product: indicator.Product{
			Name:    d.Product.Name,
			Version: d.Product.Version,
		},
		Indicators: indicators,
	}
}

func convertIndicator(i registry.APIV0Indicator) indicator.Indicator {
	thresholds := make([]indicator.Threshold, 0)
	for _, t := range i.Thresholds {
		thresholds = append(thresholds, convertThreshold(t))
	}

	return indicator.Indicator{
		Name:          i.Name,
		PromQL:        i.PromQL,
		Thresholds:    thresholds,
		Documentation: i.Documentation,
	}
}

func convertThreshold(t registry.APIV0Threshold) indicator.Threshold {
	return indicator.Threshold{
		Level:    t.Level,
		Operator: indicator.GetComparatorFromString(t.Operator),
		Value:    t.Value,
	}
}

func writeDocuments(documents []indicator.Document, outputDirectory string) {
	for _, d := range documents {
		fileBytes, _ := yaml.Marshal(AlertDocumentFrom(d))
		err := ioutil.WriteFile(fmt.Sprintf("%s/%s.yml", outputDirectory, d.Product.Name), fileBytes, 0644)
		if err != nil {
			log.Printf("error writing file: %s\n", err)
		}
	}
}
