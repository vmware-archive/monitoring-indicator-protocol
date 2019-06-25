package indicator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

func ApplyPatches(patches []Patch, documentBytes []byte) ([]byte, error) {
	var document interface{}
	err := yaml.Unmarshal(documentBytes, &document)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to unmarshal document for patching: %s", err)
	}

	for _, p := range patches {
		if MatchDocument(p.Match, documentBytes) {
			ops, err := patch.NewOpsFromDefinitions(p.Operations)
			if err != nil {
				log.Print(fmt.Errorf("failed to parse patch operations: %s", err))
				continue
			}
			for i, o := range ops {
				var tempDocument interface{}
				tempDocument, err = o.Apply(document)
				if err != nil {
					od := p.Operations[i]
					log.Print(fmt.Errorf("failed to apply patch operation %s %s: %s", od.Type, *od.Path, err))
					continue
				}
				document = tempDocument
			}
		}
	}

	patched, err := yaml.Marshal(document)
	if err != nil {
		return []byte{}, err
	}
	return patched, nil
}

func MatchDocument(criteria Match, documentBytes []byte) bool {
	reader := ioutil.NopCloser(bytes.NewReader(documentBytes))
	product, err := ProductFromYAML(reader)
	if err != nil {
		return false
	}

	if criteria.Name != nil && *criteria.Name != product.Name {
		return false
	}
	if criteria.Version != nil && *criteria.Version != product.Version {
		return false
	}

	if criteria.Metadata != nil {
		reader := ioutil.NopCloser(bytes.NewReader(documentBytes))
		metadata, err := MetadataFromYAML(reader)
		if err != nil {
			return false
		}
		if !reflect.DeepEqual(metadata, criteria.Metadata) {
			return false
		}
	}

	return true
}
