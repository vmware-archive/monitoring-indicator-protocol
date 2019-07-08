package registry

import (
	"log"
	"sync"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

type clock func() time.Time

func NewDocumentStore(timeout time.Duration, c clock) *DocumentStore {
	return &DocumentStore{
		documents:       make([]registeredDocument, 0),
		patchesBySource: make(map[string][]indicator.Patch),
		timeout:         timeout,
		getTime:         c,
	}
}

type registeredDocument struct {
	indicatorDocument indicator.Document
	registeredAt      time.Time
}

type DocumentStore struct {
	sync.RWMutex
	documents       []registeredDocument
	patchesBySource map[string][]indicator.Patch
	timeout         time.Duration
	getTime         clock
}

type PatchList struct {
	Source  string
	Patches []indicator.Patch
}

func (d *DocumentStore) UpsertDocument(doc indicator.Document) {
	d.Lock()
	defer d.Unlock()

	pos := d.getPosition(doc)

	rd := registeredDocument{
		indicatorDocument: doc,
		registeredAt:      d.getTime(),
	}

	if pos == -1 {
		d.documents = append(d.documents, rd)
	} else {
		d.documents[pos] = rd
	}
}

func (d *DocumentStore) UpsertPatches(patchList PatchList) {
	d.Lock()
	defer d.Unlock()

	d.patchesBySource[patchList.Source] = patchList.Patches

	log.Printf("registered %d patches", len(patchList.Patches))
}

func (d *DocumentStore) AllDocuments() []indicator.Document {
	d.expireDocuments()

	d.RLock()
	defer d.RUnlock()

	documents := make([]indicator.Document, 0)

	for _, doc := range d.documents {
		documents = append(documents, doc.indicatorDocument)
	}

	return documents
}

func (d *DocumentStore) FilteredDocuments(productName string) []indicator.Document {
	d.expireDocuments()

	d.RLock()
	defer d.RUnlock()

	documents := make([]indicator.Document, 0)

	for _, doc := range d.documents {
		if doc.indicatorDocument.Product.Name == productName {
			documents = append(documents, doc.indicatorDocument)
		}
	}

	return documents
}

func (d *DocumentStore) AllPatches() []indicator.Patch {
	d.RLock()
	defer d.RUnlock()

	allPatches := make([]indicator.Patch, 0)

	for _, patches := range d.patchesBySource {
		allPatches = append(allPatches, patches...)
	}

	return allPatches
}

func (d *DocumentStore) expireDocuments() {
	d.Lock()
	defer d.Unlock()

	var unexpiredDocuments []registeredDocument
	for _, doc := range d.documents {
		if !doc.registeredAt.Add(d.timeout).Before(d.getTime()) {
			unexpiredDocuments = append(unexpiredDocuments, doc)
		}
	}

	d.documents = unexpiredDocuments
}

func (d *DocumentStore) getPosition(indicatorDocument indicator.Document) int {
	for idx, doc := range d.documents {
		if doc.indicatorDocument.UID() == indicatorDocument.UID() {
			return idx
		}
	}

	return -1
}
