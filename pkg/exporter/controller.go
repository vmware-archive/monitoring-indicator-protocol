package exporter

import (
	"fmt"
	"log"
	"time"

	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/registry"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

type ControllerConfig struct {
	RegistryAPIClient registry.APIClient
	Filesystem        billy.Filesystem
	OutputDirectory   string
	UpdateFrequency   time.Duration
	DocType           string
	Converter         DocumentConverter
	Reloader          Reloader
}

type DocumentConverter func(indicator.Document) (*File, error)
type File struct {
	Name     string
	Contents []byte
}

type Reloader func() error

func NewController(c ControllerConfig) *Controller {
	if c.Filesystem == nil {
		c.Filesystem = osfs.New("/")
	}

	if c.Reloader == nil {
		c.Reloader = func() error {
			return nil
		}
	}

	return &Controller{
		Config: c,
	}
}

type Controller struct {
	Config ControllerConfig
}

func (c *Controller) Start() {
	err := c.Update()
	if err != nil {
		log.Printf("failed to update %s: %s", c.Config.DocType, err)
	}

	interval := time.NewTicker(c.Config.UpdateFrequency)
	for {
		select {
		case <-interval.C:
			err := c.Update()
			if err != nil {
				log.Printf("failed to update %s: %s", c.Config.DocType, err)
			}
		}
	}
}

func (c *Controller) Update() error {
	documents, err := c.Config.RegistryAPIClient.IndicatorDocuments()
	if err != nil {
		return fmt.Errorf("failed to fetch indicator documents, %s", err)
	}

	clearDirectory(c.Config.Filesystem, c.Config.OutputDirectory)
	writeDocuments(formatDocuments(documents), c.Config)

	return c.Config.Reloader()
}

func clearDirectory(fs billy.Filesystem, d string) {
	files, err := fs.ReadDir(d)
	if err != nil {
		log.Printf("failed to read directory %s: %s\n", d, err)
	}

	for _, f := range files {
		err = fs.Remove(fmt.Sprintf("%s/%s", d, f.Name()))
		if err != nil {
			log.Printf("failed to delete document %s: %s\n", f.Name(), err)
		}
	}
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

func writeDocuments(documents []indicator.Document, config ControllerConfig) {
	log.Printf("writing %s to %s for %d indicator documents", config.DocType, config.OutputDirectory, len(documents))

	for _, d := range documents {
		file, err := config.Converter(d)
		if err != nil {
			log.Printf("error converting document: %s\n", err)
		}

		f, err := config.Filesystem.Create(fmt.Sprintf("%s/%s", config.OutputDirectory, file.Name))
		if err != nil {
			log.Printf("error creating file: %s\n", err)
		}

		_, err = f.Write(file.Contents)
		if err != nil {
			log.Printf("error writing file: %s\n", err)
		}
	}
}
