package main_test

import (
	"code.cloudfoundry.org/indicators/pkg/indicator"
	"code.cloudfoundry.org/indicators/pkg/mtls"
	"github.com/gorilla/mux"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"testing"

	"code.cloudfoundry.org/indicators/pkg/go_test"
	"net/http"
	"os"
	"os/exec"
)

var (
	serverCert = "../../test_fixtures/leaf.pem"
	serverKey  = "../../test_fixtures/leaf.key"
	rootCACert = "../../test_fixtures/root.pem"

	clientKey  = "../../test_fixtures/client.key"
	clientCert = "../../test_fixtures/client.pem"
)

func TestIndicatorRegistryAgent(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it sends indicator documents to the registry on an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		receivedDocuments := make(chan indicator.Document, 2)

		handler := func(w http.ResponseWriter, r *http.Request) {
			g.Expect(r.URL.Query().Get("deployment")).To(Equal("abc-123"))

			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			g.Expect(err).To(Not(HaveOccurred()))

			document, err := indicator.ReadIndicatorDocument(body)
			g.Expect(err).To(Not(HaveOccurred()))

			receivedDocuments <- document
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}

		serverUrl := "127.0.0.1:34534"
		r := mux.NewRouter()
		r.HandleFunc("/v1/register", handler).Methods("POST")
		start, stop, err := mtls.NewServer(serverUrl, serverCert, serverKey, rootCACert, r)

		g.Expect(err).NotTo(HaveOccurred())

		go func() {
			err = start()
			g.Expect(err).NotTo(HaveOccurred())
		}()

		defer stop()

		println(serverUrl)

		cmd := exec.Command(binPath,
			"--documents-glob", "./test_fixtures/*/indicators.yml",
			"--registry", "https://"+serverUrl,
			"--deployment", "abc-123",
			"--tls-pem-path", clientCert,
			"--tls-key-path", clientKey,
			"--tls-root-ca-pem", rootCACert,
			"--interval", "50ms")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()

		g.Expect((<-receivedDocuments).Labels["product"]).To(Equal("job-a-product"))
		g.Expect((<-receivedDocuments).Labels["product"]).To(Equal("job-b-product"))
	})
}
