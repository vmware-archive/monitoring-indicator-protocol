package prometheus_uaa_client_test

import (
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"testing"
	"time"

	"github.com/pivotal/indicator-protocol/pkg/prometheus_uaa_client"
)

func TestUAATokenFetcher(t *testing.T) {
	t.Run("it fetches tokens from UAA and sends them to target", func(t *testing.T) {
		g := NewGomegaWithT(t)

		uaaServer := ghttp.NewServer()
		defer uaaServer.Close()

		uaaServer.AppendHandlers(
			ghttp.RespondWith(200, `{"token_type":"bearer", "access_token":"test-token"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		config := prometheus_uaa_client.UAAClientConfig{
			Insecure:        false,
			UAAHost:         uaaServer.URL(),
			UAAClientID:     "test-client",
			UAAClientSecret: "test-secret",
			Timeout:         time.Minute,
		}

		uaaClient := prometheus_uaa_client.NewUAATokenFetcher(config)

		token, err := uaaClient.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token"))
	})

	t.Run("it refreshes tokens after expiration timeout", func(t *testing.T) {
		g := NewGomegaWithT(t)

		uaaServer := ghttp.NewServer()
		defer uaaServer.Close()

		uaaServer.AppendHandlers(
			ghttp.RespondWith(200, `{"token_type":"bearer", "access_token":"test-token-1"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
			ghttp.RespondWith(200, `{"token_type":"bearer", "access_token":"test-token-2"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		config := prometheus_uaa_client.UAAClientConfig{
			Insecure:        false,
			UAAHost:         uaaServer.URL(),
			UAAClientID:     "test-client",
			UAAClientSecret: "test-secret",
			Timeout:         50 * time.Millisecond,
		}
		uaaClient := prometheus_uaa_client.NewUAATokenFetcher(config)

		token, err := uaaClient.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token-1"))

		token, err = uaaClient.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token-1"))

		time.Sleep(50 * time.Millisecond)

		token, err = uaaClient.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token-2"))
	})
}
