package registry

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func NewDocumentStore(timeout time.Duration) *DocumentStore {
	return &DocumentStore{
		documents: make([]registeredDocument, 0),
		patches:   make(map[string]registeredPatch),
		timeout:   timeout,
	}
}

type registeredDocument struct {
	indicatorDocument indicator.Document
	registeredAt      time.Time
}

type registeredPatch struct {
	indicatorPatch indicator.Patch
	registeredAt   time.Time
}

type DocumentStore struct {
	sync.RWMutex
	documents []registeredDocument
	patches   map[string]registeredPatch
	timeout   time.Duration
}

func (d *DocumentStore) UpsertDocument(doc indicator.Document) {
	d.Lock()
	defer d.Unlock()

	pos := d.getPosition(doc)

	rd := registeredDocument{
		indicatorDocument: doc,
		registeredAt:      time.Now(),
	}

	if pos == -1 {
		d.documents = append(d.documents, rd)
	} else {
		d.documents[pos] = rd
	}
}

func (d *DocumentStore) UpsertPatch(patch indicator.Patch) {
	d.Lock()
	defer d.Unlock()

	rp := registeredPatch{
		indicatorPatch: patch,
		registeredAt:   time.Now(),
	}

	d.patches[patch.Origin] = rp
	logPatchInsert(rp)
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

func (d *DocumentStore) AllPatches() []indicator.Patch {
	d.RLock()
	defer d.RUnlock()

	patches := make([]indicator.Patch, 0)

	for _, patch := range d.patches {
		patches = append(patches, patch.indicatorPatch)
	}

	return patches
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
		if reflect.DeepEqual(doc.indicatorDocument.Metadata, indicatorDocument.Metadata) && doc.indicatorDocument.Product.Name == indicatorDocument.Product.Name {
			return idx
		}
	}

	return -1
}

func logPatchInsert(rp registeredPatch) {
	logLine := strings.Builder{}
	logLine.Write([]byte("registered patch for"))
	if rp.indicatorPatch.Match.Name != nil {
		logLine.WriteString(" name: " + *rp.indicatorPatch.Match.Name)
	}
	if rp.indicatorPatch.Match.Version != nil {
		logLine.WriteString(" version: " + *rp.indicatorPatch.Match.Version)
	}
	if rp.indicatorPatch.Match.Metadata != nil {
		logLine.WriteString(fmt.Sprintf(" metadata: %v", rp.indicatorPatch.Match.Metadata))
	}
	log.Println(logLine.String())
}
