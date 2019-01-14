package configuration

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"

	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func Read(configFile string) ([]indicator.Patch, error) {
	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s\n", err)
	}

	var f file
	err = yaml.Unmarshal(fileBytes, &f)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %s\n", err)
	}

	var patches []indicator.Patch
	for _, patchPath := range f.Sources {
		patch, err := indicator.ReadPatchFile(patchPath)
		if err != nil {
			log.Printf("failed to read patch %s from config file %s: %s\n", patchPath, configFile, err)
			continue
		}

		patches = append(patches, patch)
	}

	return patches, nil
}

type file struct {
	Sources []string `yaml:"sources"`
}
