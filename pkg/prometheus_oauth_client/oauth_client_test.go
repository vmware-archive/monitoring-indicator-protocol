package prometheus_oauth_client_test

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_oauth_client"
)

func TestTokenFetcher(t *testing.T) {
	t.Run("it fetches tokens from the OAuth server and sends them to target", func(t *testing.T) {
		g := NewGomegaWithT(t)

		oauthServer := ghttp.NewServer()
		defer oauthServer.Close()

		oauthServer.AppendHandlers(
			ghttp.RespondWith(200, `{"token_type":"bearer", "access_token":"test-token"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		config := prometheus_oauth_client.OAuthClientConfig{
			Insecure:          false,
			OAuthServer:       oauthServer.URL(),
			OAuthClientID:     "test-client",
			OAuthClientSecret: "test-secret",
			Timeout:           time.Minute,
			Clock:             time.Now,
		}

		client := prometheus_oauth_client.NewTokenFetcher(config)

		token, err := client.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token"))
	})

	t.Run("it refreshes tokens after expiration timeout", func(t *testing.T) {
		g := NewGomegaWithT(t)

		oauthServer := ghttp.NewServer()
		defer oauthServer.Close()

		oauthServer.AppendHandlers(
			ghttp.RespondWith(200, `{"token_type":"bearer", "access_token":"test-token-1"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
			ghttp.RespondWith(200, `{"token_type":"bearer", "access_token":"test-token-2"}`, map[string][]string{"Content-Type:": {"application/json;charset=UTF-8"}}),
		)

		theTime := time.Now()
		config := prometheus_oauth_client.OAuthClientConfig{
			Insecure:          false,
			OAuthServer:       oauthServer.URL(),
			OAuthClientID:     "test-client",
			OAuthClientSecret: "test-secret",
			Timeout:           time.Hour,
			Clock:             func() time.Time { return theTime },
		}
		client := prometheus_oauth_client.NewTokenFetcher(config)

		token, err := client.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token-1"))

		token, err = client.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token-1"))

		theTime = theTime.Add(time.Hour).Add(time.Millisecond)

		token, err = client.GetClientToken()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(token).To(Equal("bearer test-token-2"))
	})
}
