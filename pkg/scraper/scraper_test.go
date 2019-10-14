package scraper_test

import (
	"io/ioutil"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/scraper"
)

func TestMakeApiClients(t *testing.T) {
	g := NewGomegaWithT(t)
	caCertBytes, err := ioutil.ReadFile("../../test_fixtures/ca.pem")
	g.Expect(err).ToNot(HaveOccurred())
	clientKeyBytes, err := ioutil.ReadFile("../../test_fixtures/client.key")
	g.Expect(err).ToNot(HaveOccurred())
	clientCertBytes, err := ioutil.ReadFile("../../test_fixtures/client.pem")
	g.Expect(err).ToNot(HaveOccurred())

	validConfig := scraper.RemoteScrapeConfig{
		SourceName:   "foo",
		ServerName:   "localhost",
		RegistryAddr: "https://localhost:1234",
		CaCert:       string(caCertBytes),
		ClientCreds: scraper.ClientCreds{
			ClientKey:  string(clientKeyBytes),
			ClientCert: string(clientCertBytes),
		},
	}

	t.Run("Given empty or nil, returns empty", func(t *testing.T) {
		g := NewGomegaWithT(t)

		nilResult, errs := scraper.MakeApiClients(nil)
		g.Expect(errs).To(HaveLen(0))
		g.Expect(nilResult).To(BeEmpty())

		emptyResult, errs := scraper.MakeApiClients([]scraper.RemoteScrapeConfig{})
		g.Expect(errs).To(HaveLen(0))
		g.Expect(emptyResult).To(BeEmpty())
	})

	t.Run("Can construct api clients for remote foundations", func(t *testing.T) {
		g := NewGomegaWithT(t)
		configs := []scraper.RemoteScrapeConfig{validConfig}
		apiClients, errs := scraper.MakeApiClients(configs)
		g.Expect(apiClients).To(HaveLen(1))
		g.Expect(errs).To(HaveLen(0))

		// We want to test the contents of the api client, like the certs, but those fields are
		// unexported. This behavior is tested in the integration tests, because the connections couldn't
		// be made if the certs were not being passed through.
	})

	t.Run("Errors if any of the mTLS configs are invalid", func(t *testing.T) {
		g := NewGomegaWithT(t)

		configs := []scraper.RemoteScrapeConfig{
			validConfig,
			{
				SourceName:   "source",
				ServerName:   "server",
				RegistryAddr: "localhost:123",
				CaCert:       "sdaerq23waezrfzr",
				ClientCreds:  scraper.ClientCreds{},
			},
		}
		_, errs := scraper.MakeApiClients(configs)

		g.Expect(errs).To(HaveLen(1))
		g.Expect(errs[0].Error()).Should(ContainSubstring("mTLS remote client"))
	})
}

type mockLocalApiClient struct {
}

func (c mockLocalApiClient) Register(document v1.IndicatorDocument) error {
	return nil
}

type mockRemoteApiClient struct {
	forwardDocumentInvocations uint64
	mu                         sync.Mutex
}

func (c *mockRemoteApiClient) ForwardDocumentsTo(destinationApiClient scraper.DocumentRegistrar) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.forwardDocumentInvocations += 1
}

func (c *mockRemoteApiClient) getInvocations() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.forwardDocumentInvocations
}

func TestRunLoop(t *testing.T) {
	t.Run("Should forward each remote's documents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		localApiClient := mockLocalApiClient{}
		remoteApiClient1 := &mockRemoteApiClient{}
		remoteApiClient2 := &mockRemoteApiClient{}

		remoteApiClients := []scraper.RemoteFoundationApiClient{
			remoteApiClient1, remoteApiClient2,
		}

		killScraper := scraper.RunLoop(100*time.Millisecond, remoteApiClients, localApiClient)

		g.Eventually(func() uint64 {
			remoteApiClient1.mu.Lock()
			defer remoteApiClient1.mu.Unlock()
			return remoteApiClient1.forwardDocumentInvocations
		}, 3*time.Second).Should(BeNumerically(">", 2))

		killScraper()

		g.Expect(remoteApiClient1.getInvocations()).To(Equal(remoteApiClient2.getInvocations()))
	})
}
