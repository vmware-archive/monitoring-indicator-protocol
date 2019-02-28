package exporter

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
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
	outputDir := c.Config.OutputDirectory
	fs := c.Config.Filesystem

	err := fs.MkdirAll(outputDir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %s\n", outputDir, err)
	}

	apiDocuments, err := c.Config.RegistryAPIClient.IndicatorDocuments()
	documents := formatDocuments(apiDocuments)
	if err != nil {
		return fmt.Errorf("failed to fetch indicator documents, %s", err)
	}

	err = clearDirectory(fs, outputDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %s\n", outputDir, err)
	}
	writeDocuments(documents, c.Config)

	return c.Config.Reloader()
}

func formatDocuments(documents []registry.APIV0Document) []indicator.Document {
	formattedDocuments := make([]indicator.Document, 0)
	for _, d := range documents {
		formattedDocuments = append(formattedDocuments, registry.ToIndicatorDocument(d))
	}

	return formattedDocuments
}

func clearDirectory(fs billy.Filesystem, d string) error {
	files, err := fs.ReadDir(d)
	if err != nil {
		return err
	}

	for _, f := range files {
		err = fs.Remove(fmt.Sprintf("%s/%s", d, f.Name()))
		if err != nil {
			log.Printf("failed to delete document %s: %s\n", f.Name(), err)
		}
	}

	return nil
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
