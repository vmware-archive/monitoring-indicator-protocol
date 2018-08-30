package main_test

import (
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"testing"

	"code.cloudfoundry.org/cf-indicators/pkg/go_test"
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os"
	"os/exec"
)

func TestIndicatorRegistryAgent(t *testing.T) {
	g := NewGomegaWithT(t)

	binPath, err := go_test.Build("./")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it sends indicator documents to the registry on an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

		registryServer := ghttp.NewServer()
		defer registryServer.Close()

		receivedDocuments := make(chan indicator.Document, 2)
		registryServer.RouteToHandler("POST", "/v1/register", func(w http.ResponseWriter, r *http.Request) {
			g.Expect(r.URL.Query().Get("deployment")).To(Equal("abc-123"))

			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			g.Expect(err).To(Not(HaveOccurred()))

			document, err := indicator.ReadIndicatorDocument(body)
			g.Expect(err).To(Not(HaveOccurred()))

			receivedDocuments <- document
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})

		println(registryServer.URL())

		cmd := exec.Command(binPath,
			"--documents-glob", "./test_fixtures/*/indicators.yml",
			"--registry", registryServer.URL(),
			"--deployment", "abc-123",
			"--interval", "50ms")

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)

		g.Expect(err).ToNot(HaveOccurred())
		defer session.Kill()

		g.Expect((<-receivedDocuments).Product).To(Equal("job-a-product"))
		g.Expect((<-receivedDocuments).Product).To(Equal("job-b-product"))
	})
}
