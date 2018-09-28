package registry

import (
	"reflect"
	"sync"
	"time"

	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func NewDocumentStore(timeout time.Duration) *DocumentStore {
	return &DocumentStore{
		documents: make([]Document, 0),
		timeout:   timeout,
	}
}

type DocumentStore struct {
	sync.RWMutex
	documents []Document
	timeout   time.Duration
}

func (d *DocumentStore) Upsert(labels map[string]string, indicators []indicator.Indicator) {
	d.Lock()
	defer d.Unlock()

	pos := d.getPosition(labels)

	if pos == -1 {
		d.documents = append(d.documents, Document{Indicators: indicators, Labels: labels, registrationTimestamp: time.Now()})
	} else {
		d.documents[pos] = Document{Indicators: indicators, Labels: labels, registrationTimestamp: time.Now()}
	}
}

func (d *DocumentStore) All() []Document {
	d.expireDocuments()

	d.RLock()
	defer d.RUnlock()

	return d.documents
}

func (d *DocumentStore) expireDocuments() {
	d.Lock()
	defer d.Unlock()

	var unexpiredDocuments []Document
	for _, doc := range d.documents {
		if !doc.registrationTimestamp.Add(d.timeout).Before(time.Now()) {
			unexpiredDocuments = append(unexpiredDocuments, doc)
		}
	}

	d.documents = unexpiredDocuments
}

func (d *DocumentStore) getPosition(labels map[string]string) int {
	for idx, doc := range d.documents {
		if reflect.DeepEqual(doc.Labels, labels) {
			return idx
		}
	}
	return -1
}
