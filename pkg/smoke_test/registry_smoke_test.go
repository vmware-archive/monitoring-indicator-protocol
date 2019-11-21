package smoke_test

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/tls_config"
)

var (
	clientKey, clientCert, rootCACert, registryUrl, registryCommonName string
	timeoutInterval                                                    time.Duration
	client                                                             *http.Client
	registryClient                                                     *registry.RegistryApiClient
)

func TestMain(m *testing.M) {
	flag.StringVar(&clientKey, "tls-key-path", "", "")
	flag.StringVar(&clientCert, "tls-pem-path", "", "")
	flag.StringVar(&rootCACert, "tls-root-ca-pem", "", "")
	flag.StringVar(&registryUrl, "registry", "", "")
	flag.StringVar(&registryCommonName, "tls-server-cn", "", "")
	flag.DurationVar(&timeoutInterval, "interval", 2*time.Minute, "")
	flag.Parse()
	os.Exit(m.Run())
}

func TestIndicatorRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if clientKey == "" {
		log.Panic("Oh no! tls-key-path not provided")
	}
	if clientCert == "" {
		log.Panic("Oh no! tls-pem-path not provided")
	}
	if rootCACert == "" {
		log.Panic("Oh no! tls-root-ca-pem not provided")
	}
	if registryUrl == "" {
		log.Panic("Oh no! registry not provided")
	}
	if registryCommonName == "" {
		log.Panic("Oh no! tls-server-cn not provided")
	}
	tlsConfig, err := tls_config.NewClientConfig(clientCert, clientKey, rootCACert, registryCommonName)
	if err != nil {
		log.Panic(fmt.Sprintf("Oh no! couldn't create a tls config: %s", err))
	}

	client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   tlsConfig,
		},
	}

	registryClient = registry.NewAPIClient(registryUrl, client)
	t.Run("it returns indicator documents", func(t *testing.T) {
		g := NewGomegaWithT(t)

		g.Eventually(retrieveIndicatorDocs, timeoutInterval).ShouldNot(HaveLen(0))

		apiv0Documents, _ := registryClient.IndicatorDocuments()

		g.Expect(containsProductName(apiv0Documents, "indicator-protocol-registry")).To(BeTrue())
		log.Print("Smoke test success!")
	})
}

func retrieveIndicatorDocs() []registry.APIDocumentResponse {
	apiv0Documents, e := registryClient.IndicatorDocuments()
	if e != nil {
		panic(fmt.Errorf("oh no! Could not query the registry: %s", e))
	}
	return apiv0Documents
}

func containsProductName(documents []registry.APIDocumentResponse, name string) bool {
	for _, listItem := range documents {
		if listItem.Spec.Product.Name == name {
			return true
		}
	}

	return false
}
