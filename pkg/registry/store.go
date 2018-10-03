package registry

import (
	"reflect"
	"sync"
	"time"

	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func NewDocumentStore(timeout time.Duration) *DocumentStore {
	return &DocumentStore{
		documents: make([]registeredDocument, 0),
		timeout:   timeout,
	}
}

type registeredDocument struct {
	indicatorDocument indicator.Document
	registeredAt      time.Time
}

type DocumentStore struct {
	sync.RWMutex
	documents []registeredDocument
	timeout   time.Duration
}

func (d *DocumentStore) Upsert(doc indicator.Document) {
	d.Lock()
	defer d.Unlock()

	pos := d.getPosition(doc)

	if pos == -1 {
		d.documents = append(d.documents, registeredDocument{indicatorDocument: doc, registeredAt: time.Now()})
	} else {
		d.documents[pos] = registeredDocument{indicatorDocument: doc, registeredAt: time.Now()}
	}
}

func (d *DocumentStore) All() []indicator.Document {
	d.expireDocuments()

	d.RLock()
	defer d.RUnlock()

	documents := make([]indicator.Document, 0)

	for _, doc := range d.documents {
		documents = append(documents, doc.indicatorDocument)
	}

	return documents
}

func (d *DocumentStore) expireDocuments() {
	d.Lock()
	defer d.Unlock()

	var unexpiredDocuments []registeredDocument
	for _, doc := range d.documents {
		if !doc.registeredAt.Add(d.timeout).Before(time.Now()) {
			unexpiredDocuments = append(unexpiredDocuments, doc)
		}
	}

	d.documents = unexpiredDocuments
}

func (d *DocumentStore) getPosition(indicatorDocument indicator.Document) int {
	for idx, doc := range d.documents {
		if reflect.DeepEqual(doc.indicatorDocument.Metadata, indicatorDocument.Metadata) && doc.indicatorDocument.Product == indicatorDocument.Product {
			return idx
		}
	}

	return -1
}
