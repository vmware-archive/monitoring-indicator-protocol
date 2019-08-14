package exporter

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

type APIClient interface {
	IndicatorDocuments() ([]registry.APIDocumentResponse, error)
}

type ControllerConfig struct {
	RegistryAPIClient APIClient
	Filesystem        billy.Filesystem
	OutputDirectory   string
	UpdateFrequency   time.Duration
	DocType           string
	Converter         DocumentConverter
	Reloader          Reloader
}

type DocumentConverter func(v1.IndicatorDocument) (*File, error)
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
		return errors.New("failed to create output directory")
	}

	apiDocuments, err := c.Config.RegistryAPIClient.IndicatorDocuments()
	if err != nil {
		return errors.New("failed to fetch indicator documents")
	}

	err = clearDirectory(fs, outputDir)
	if err != nil {
		return fmt.Errorf("failed to delete contents of directory: %s", err)
	}
	writeDocuments(apiDocuments, c.Config)

	return c.Config.Reloader()
}

func clearDirectory(fs billy.Filesystem, d string) error {
	files, err := fs.ReadDir(d)
	if err != nil {
		return errors.New("failed to read directory")
	}

	// Note that this will not return an error if some documents are not deleted,
	// it will just print a log statement.
	for _, f := range files {
		err = fs.Remove(fmt.Sprintf("%s/%s", d, f.Name()))
		if err != nil {
			log.Print("failed to delete a document")
		}
	}

	return nil
}

func writeDocuments(documents []registry.APIDocumentResponse, config ControllerConfig) {
	log.Printf("writing %d indicator documents to output directory", len(documents))

	for _, d := range documents {
		file, err := config.Converter(registry.ToIndicatorDocument(d))
		if err != nil {
			log.Print("error converting document")
		}

		f, err := config.Filesystem.Create(fmt.Sprintf("%s/%s", config.OutputDirectory, file.Name))
		if err != nil {
			log.Print("error creating file")
		}

		_, err = f.Write(file.Contents)
		if err != nil {
			log.Print("error writing file")
		}
	}
}
