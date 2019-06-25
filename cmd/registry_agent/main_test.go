package main_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
)

var (
	serverCert = "../../test_fixtures/server.pem"
	serverKey  = "../../test_fixtures/server.key"
	rootCACert = "../../test_fixtures/ca.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestIndicatorRegistryAgent(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it sends indicator documents to the registry on an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		receivedDocuments := make(chan indicator.Document, 2)

		handler := func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			g.Expect(err).To(Not(HaveOccurred()))

			reader := ioutil.NopCloser(bytes.NewReader(body))
			document, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).To(Not(HaveOccurred()))

			receivedDocuments <- document
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}

		serverUrl := "127.0.0.1:34534"
		r := mux.NewRouter()
		r.HandleFunc("/v1/register", handler).Methods("POST")

		tlsConfig, err := mtls.NewServerConfig(rootCACert)
		g.Expect(err).NotTo(HaveOccurred())

		server := &http.Server{
			Addr:      serverUrl,
			Handler:   r,
			TLSConfig: tlsConfig,
		}

		start := func() error { return server.ListenAndServeTLS(serverCert, serverKey) }
		stop := func() error { return server.Close() }

		done := make(chan struct{})
		defer func() {
			_ = stop()
			<-done
		}()
		go func() {
			defer close(done)
			_ = start()
		}()

		cmd := exec.Command(
			binPath,
			"--documents-glob", "./test_fixtures/*/indicators.yml",
			"--registry", "https://"+serverUrl,
			"--tls-pem-path", clientCert,
			"--tls-key-path", clientKey,
			"--tls-root-ca-pem", rootCACert,
			"--tls-server-cn", "localhost",
			"--interval", "50ms",
		)

		session, err := gexec.Start(cmd, nil, nil)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()

		g.Expect((<-receivedDocuments).Product.Name).To(Equal("job-a-product"))
		g.Expect((<-receivedDocuments).Product.Name).To(Equal("job-b-product"))
	})
}
