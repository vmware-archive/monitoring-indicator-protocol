package registry_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
)

func TestServingMetrics(t *testing.T) {
	g := NewGomegaWithT(t)

	conf := registry.WebServerConfig{
		Address: "localhost:12345",
	}
	start, stop := registry.NewWebServer(conf)
	defer stop()

	go start()

	var resp *http.Response
	f := func() error {
		var err error
		resp, err = http.Get(fmt.Sprintf("http://%s/metrics", conf.Address))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("invalid status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
	g.Eventually(f).ShouldNot(HaveOccurred())
}

func TestRegisterAndServeDocuments(t *testing.T) {
	g := NewGomegaWithT(t)

	conf := registry.WebServerConfig{
		Address:       "localhost:12346",
		DocumentStore: registry.NewDocumentStore(time.Second, time.Now),
		StatusStore:   status_store.New(time.Now),
	}
	start, stop := registry.NewWebServer(conf)
	defer stop()

	go start()

	var resp *http.Response
	f := func() error {
		var err error
		resp, err = http.Post(
			fmt.Sprintf("http://%s/v1/register", conf.Address),
			"application/json",
			bytes.NewReader([]byte(document)),
		)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("invalid status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
	g.Eventually(f).ShouldNot(HaveOccurred())

	resp, err := http.Get(fmt.Sprintf("http://%s/v1/indicator-documents", conf.Address))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, err := ioutil.ReadAll(resp.Body)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(body).To(MatchJSON(fmt.Sprintf("[%s]", document)))
}

func TestWritingAndReadingStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	conf := registry.WebServerConfig{
		Address:       "localhost:12347",
		DocumentStore: registry.NewDocumentStore(time.Second, time.Now),
		StatusStore:   status_store.New(time.Now),
	}
	start, stop := registry.NewWebServer(conf)
	defer stop()

	go start()

	var resp *http.Response
	f := func() error {
		var err error
		resp, err = http.Post(
			fmt.Sprintf("http://%s/v1/register", conf.Address),
			"application/json",
			bytes.NewReader([]byte(document)),
		)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("invalid status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
	g.Eventually(f).ShouldNot(HaveOccurred())

	// make our status update request
	const documentUID = `my-product-a-a902332065d69c1787f419e235a1f1843d98c884`
	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/indicator-documents/%s/bulk_status", conf.Address, documentUID),
		"application/json",
		bytes.NewReader([]byte(statusRequest)),
	)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	// get document
	resp, err = http.Get(fmt.Sprintf("http://%s/v1/indicator-documents", conf.Address))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, err := ioutil.ReadAll(resp.Body)
	g.Expect(err).ToNot(HaveOccurred())

	var documents []registry.APIV0Document
	err = json.Unmarshal(body, &documents)
	g.Expect(err).ToNot(HaveOccurred())

	var (
		indie1Status *string
		indie2Status *string
	)
	for _, doc := range documents {
		if doc.UID == documentUID {
			for _, ind := range doc.Indicators {
				switch ind.Name {
				case "indie1":
					indie1Status = ind.Status.Value
				case "indie2":
					indie2Status = ind.Status.Value
				}
			}
		}
	}

	g.Expect(indie1Status).To(Equal(strPtr("critical")))
	g.Expect(indie2Status).To(Equal(strPtr("warning")))
}

const (
	statusRequest = `[
			{
				"name": "indie1",
				"status": "critical"
			},
			{
				"name": "indie2",
				"status": "warning"
			}
		]`
	document = `{
		  "apiVersion": "v1alpha1",
		  "uid": "my-product-a-a902332065d69c1787f419e235a1f1843d98c884",
		  "product": {
			"name": "my-product-a",
			"version": "1"
		  },
		  "metadata": {
			"deployment": "abc-123"
		  },
		  "indicators": [
			{
			  "name": "indie1",
			  "promql": "promql1",
			  "thresholds": [],
			  "alert": {
				"for": "5m",
				"step": "10s"
			  },
			  "serviceLevel": null,
			  "presentation": {
				"chartType": "step",
				"currentValue": false,
				"frequency": 0,
				"labels": [],
				"units": ""
			  },
			  "status": null
			},
			{
			  "name": "indie2",
			  "promql": "promql2",
			  "thresholds": [],
			  "alert": {
				"for": "5m",
				"step": "10s"
			  },
			  "serviceLevel": {
				"objective": 99.99
			  },
			  "presentation": {
				"chartType": "status",
				"currentValue": false,
				"frequency": 0,
				"labels": [],
				"units": "nanoseconds"
			  },
			  "status": null
			}
		  ],
		  "layout": {
			"title": "",
			"description": "",
			"sections": [],
			"owner": ""
		  }
		}`
)
