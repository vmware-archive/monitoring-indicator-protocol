package registry_proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry_proxy"
)

func TestRegistryProxy(t *testing.T) {
	t.Run("it broadcast POST requests to every registry", func(t *testing.T) {
		g := NewGomegaWithT(t)
		rw := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/foo", nil)

		var (
			localCalled, remoteCalled   bool
			localURLPath, remoteURLPath string
		)

		handlerFunc := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			localCalled = true
			localURLPath = r.URL.Path
		})
		h := registry_proxy.NewHandler(handlerFunc, []http.Handler{http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			remoteCalled = true
			remoteURLPath = r.URL.Path
		})})

		h.ServeHTTP(rw, r)

		g.Expect(localCalled).To(BeTrue())
		g.Expect(localURLPath).To(Equal("/foo"))
		g.Expect(remoteCalled).To(BeTrue())
		g.Expect(remoteURLPath).To(Equal("/backend/foo"))
	})

	t.Run("it broadcasts requests in parallel", func(t *testing.T) {
		g := NewGomegaWithT(t)
		rw := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/foo", nil)

		var (
			firstHandlerRan, secondHandlerRan bool
		)

		done1 := make(chan bool)
		done2 := make(chan bool)

		handlerFunc := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {})
		h := registry_proxy.NewHandler(
			handlerFunc,
			[]http.Handler{
				http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					<-done1
					firstHandlerRan = true
					done2 <- true
				}),
				http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					secondHandlerRan = true
					done1 <- true
				}),
			},
		)

		go h.ServeHTTP(rw, r)
		<-done2

		g.Expect(firstHandlerRan).To(BeTrue())
		g.Expect(firstHandlerRan).To(BeTrue())
		g.Expect(secondHandlerRan).To(BeTrue())
	})

	t.Run("it only sends to local registry", func(t *testing.T) {
		g := NewGomegaWithT(t)

		testCases := map[string]*http.Request{
			"GET request":                        httptest.NewRequest("GET", "/foo", nil),
			"GET request with /backend/ prefix":  httptest.NewRequest("GET", "/backend/foo", nil),
			"POST request with /backend/ prefix": httptest.NewRequest("POST", "/backend/foo", nil),
		}

		for name, req := range testCases {
			t.Run(name, func(t *testing.T) {
				rw := httptest.NewRecorder()

				var (
					localCalled, remoteCalled bool
					localURLPath              string
				)
				h := registry_proxy.NewHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					localCalled = true
					localURLPath = req.URL.Path
				}), []http.Handler{http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					remoteCalled = true
				})})

				h.ServeHTTP(rw, req)
				g.Expect(localCalled).To(BeTrue())
				g.Expect(localURLPath).To(Equal("/foo"))
				g.Expect(remoteCalled).To(BeFalse())
			})
		}
	})
}
