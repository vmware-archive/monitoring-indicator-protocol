package registry

import (
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"sync"
)

type Document struct {
	Labels     map[string]string     `json:"labels"`
	Indicators []indicator.Indicator `json:"indicators"`
}

type Indicator struct {
	indicator.Indicator
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

func (d *DocumentStore) Insert(labels map[string]string, indicators []indicator.Indicator) {
	d.Lock()
	defer d.Unlock()

	d.documents = append(d.documents, Document{Indicators: indicators, Labels: labels})
}

func (d *DocumentStore) All() []Document {
	d.RLock()
	defer d.RUnlock()

	return d.documents
}
