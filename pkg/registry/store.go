package registry

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pivotal/indicator-protocol/pkg/indicator"
)

func NewDocumentStore(timeout time.Duration) *DocumentStore {
	return &DocumentStore{
		documents:       make([]registeredDocument, 0),
		patchesBySource: make(map[string][]indicator.Patch),
		timeout:         timeout,
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
		registeredAt:      time.Now(),
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

	for _, p := range patchList.Patches {
		logPatchInsert(p)
	}
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

func logPatchInsert(p indicator.Patch) {
	logLine := strings.Builder{}
	logLine.Write([]byte("registered patch for"))
	if p.Match.Name != nil {
		logLine.WriteString(" name: " + *p.Match.Name)
	}
	if p.Match.Version != nil {
		logLine.WriteString(" version: " + *p.Match.Version)
	}
	if p.Match.Metadata != nil {
		logLine.WriteString(fmt.Sprintf(" metadata: %v", p.Match.Metadata))
	}
	log.Println(logLine.String())
}
