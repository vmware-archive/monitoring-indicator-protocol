package e2e_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
)

const (
	rootCaCert = "../../test_fixtures/ca.pem"

	serverCert = "../../test_fixtures/server.pem"
	serverKey  = "../../test_fixtures/server.key"

	clientCert = "../../test_fixtures/client.pem"
	clientKey  = "../../test_fixtures/client.key"

	cfAuthProxyPort = "15432"
	registryPort    = "23819"
	tlsProxyPort    = "12911"

	// This must be uaa.sys.…, not …cf-app.com/uaa, otherwise everything
	// will fail. You also must be targeting madlamp to run this locally.
	uaaUrl = "https://uaa.madlamp.cf-denver.com"
)

var (
	getDocumentsUrlString = fmt.Sprintf("https://localhost:%s/v1/indicator-documents", cfAuthProxyPort)

	tlsConfig, _ = mtls.NewSingleAuthClientConfig(rootCaCert, "localhost")
	client       = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
)

var bearerRE = regexp.MustCompile(`(?i)^bearer\s+`)

func trimBearer(authToken string) string {
	trimmedToken := bearerRE.ReplaceAllString(authToken, "")
	return trimmedToken
}

func TestCfAuthProxy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	
	tokenBytes, err := exec.Command("cf", "oauth-token").Output()
	if err != nil {
		t.Fatal(err)
	}
	tokenString := trimBearer(strings.TrimSuffix(string(tokenBytes), "\n"))
	kill, err := startSession(uaaUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer kill()

	t.Run("UAA Admin can retrieve documents using token query param", func(t *testing.T) {
		g := NewGomegaWithT(t)
		getDocumentsUrl := fmt.Sprintf("%s?token=%s", getDocumentsUrlString, tokenString)

		response, err := client.Get(getDocumentsUrl)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response.StatusCode).To(Equal(http.StatusOK),
			"Could not retrieve documents. Make sure you have targeted madlamp.")
		response1Bytes, _ := ioutil.ReadAll(response.Body)
		g.Expect(response1Bytes).To(MatchJSON([]byte("[]")))
	})

	t.Run("UAA Admin can retrieve documents using header to provide token", func(t *testing.T) {
		g := NewGomegaWithT(t)

		header := http.Header{}
		header.Add(
			"Authorization",
			fmt.Sprintf("Bearer %s", tokenString),
		)

		getDocumentsUrl, _ := url.Parse(getDocumentsUrlString)

		request := &http.Request{
			Method: http.MethodGet,
			URL:    getDocumentsUrl,
			Header: header,
		}
		response, err := client.Do(request)

		g.Expect(err).NotTo(HaveOccurred(),
			"Could not retrieve documents. Make sure you have targeted madlamp.")
		g.Expect(response.StatusCode).To(Equal(http.StatusOK))
		response1Bytes, _ := ioutil.ReadAll(response.Body)
		g.Expect(response1Bytes).To(MatchJSON([]byte(fmt.Sprintf("[]"))))
	})

	t.Run("If not a UAA Admin, can't get documents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		response, err := client.Get(getDocumentsUrlString)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response.StatusCode).To(Equal(http.StatusForbidden))
	})
}

func startSession(uaaUrl string) (func(), error) {

	binPath, err := go_test.Build("../../cmd/cf_auth_proxy", "-race")
	cmd := exec.Command(
		binPath,
		"--port", cfAuthProxyPort,
		"--tls-client-pem-path", clientCert,
		"--tls-client-key-path", clientKey,
		"--tls-root-ca-pem", rootCaCert,
		"--tls-key-path", serverKey,
		"--tls-pem-path", serverCert,
		"--uaa-addr", uaaUrl,
		"--registry-addr", "https://localhost:"+tlsProxyPort,
	)
	cfAuthProxySession, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
	binPath, _ = go_test.Build("../../cmd/registry_proxy", "-race")
	cmd = exec.Command(
		binPath,
		"--port", tlsProxyPort,
		"--tls-pem-path", serverCert,
		"--tls-client-pem-path", clientCert,
		"--tls-key-path", serverKey,
		"--tls-client-key-path", clientKey,
		"--tls-root-ca-pem", rootCaCert,
		"--local-registry-addr", fmt.Sprintf("localhost:%s", registryPort),
	)
	tlsProxySession, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	binPath, _ = go_test.Build("../../cmd/registry", "-race")

	cmd = exec.Command(
		binPath,
		"--port", registryPort,
	)
	registrySession, _ := gexec.Start(cmd, os.Stdout, os.Stderr)

	kill := func() {
		cfAuthProxySession.Kill()
		tlsProxySession.Kill()
		registrySession.Kill()
	}
	err = go_test.WaitForTCPServer(fmt.Sprintf("localhost:%s", cfAuthProxyPort), 3*time.Second)
	if err != nil {
		return kill, err
	}
	err = go_test.WaitForTCPServer(fmt.Sprintf("localhost:%s", tlsProxyPort), 3*time.Second)
	if err != nil {
		return kill, err
	}
	err = go_test.WaitForTCPServer(fmt.Sprintf("localhost:%s", registryPort), 3*time.Second)
	if err != nil {
		return kill, err
	}

	return kill, nil
}
