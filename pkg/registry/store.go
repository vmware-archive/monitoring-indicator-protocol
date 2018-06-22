package registry

import (
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"sync"
	"reflect"
)

type Document struct {
	Labels     map[string]string
	Indicators []indicator.Indicator
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: make([]Document, 0),
	}
}

type DocumentStore struct {
	sync.RWMutex
	documents []Document
}

func (d *DocumentStore) Upsert(labels map[string]string, indicators []indicator.Indicator) {
	d.Lock()
	defer d.Unlock()

	pos := d.getPosition(labels)

	if pos == -1 {
		d.documents = append(d.documents, Document{Indicators: indicators, Labels: labels})
	} else {
		d.documents[pos] = Document{Indicators: indicators, Labels: labels}
	}
}

func (d *DocumentStore) All() []Document {
	d.RLock()
	defer d.RUnlock()

	return d.documents
}

func (d *DocumentStore) getPosition(labels map[string]string) int {
	for idx, doc := range d.documents {
		if reflect.DeepEqual(doc.Labels, labels) {
			return idx
		}
	}
	return -1
}
