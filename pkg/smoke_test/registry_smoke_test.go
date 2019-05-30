package smoke_test

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

var (
	clientKey, clientCert, rootCACert, registryUrl, registryCommonName *string
	timeoutInterval                                                    *time.Duration
	client                                                             *http.Client
	registryClient                                                     *registry.RegistryApiClient
)

func init() {
	clientKey = flag.String("tls-key-path", "", "")
	clientCert = flag.String("tls-pem-path", "", "")
	rootCACert = flag.String("tls-root-ca-pem", "", "")
	registryUrl = flag.String("registry", "", "")
	registryCommonName = flag.String("tls-server-cn", "", "")
	timeoutInterval = flag.Duration("interval", 2*time.Minute, "")
	flag.Parse()
	if *clientKey == "" {
		log.Panic("Oh no! tls-key-path not provided")
	}
	if *clientCert == "" {
		log.Panic("Oh no! tls-pem-path not provided")
	}
	if *rootCACert == "" {
		log.Panic("Oh no! tls-root-ca-pem not provided")
	}
	if *registryUrl == "" {
		log.Panic("Oh no! registry not provided")
	}
	if *registryCommonName == "" {
		log.Panic("Oh no! tls-server-cn not provided")
	}
	tlsConfig, err := mtls.NewClientConfig(*clientCert, *clientKey, *rootCACert, *registryCommonName)
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

	registryClient = registry.NewAPIClient(*registryUrl, client)
}

func TestIndicatorRegistry(t *testing.T) {
	t.Run("it returns indicator documents", func(t *testing.T) {
		g := NewGomegaWithT(t)

		g.Eventually(retrieveIndicatorDocs, *timeoutInterval).ShouldNot(HaveLen(0))

		apiv0Documents, _ := registryClient.IndicatorDocuments()

		g.Expect(containsProductName(apiv0Documents, "indicator-protocol-registry")).To(BeTrue())
		log.Print("Smoke test success!")
	})
}

func retrieveIndicatorDocs() []registry.APIV0Document {
	apiv0Documents, e := registryClient.IndicatorDocuments()
	if e != nil {
		panic(fmt.Errorf("oh no! Could not query the registry: %s", e))
	}
	return apiv0Documents
}

func containsProductName(documents []registry.APIV0Document, name string) bool {
	for _, listItem := range documents {
		if listItem.Product.Name == name {
			return true
		}
	}

	return false
}
