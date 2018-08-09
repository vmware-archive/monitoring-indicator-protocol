package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type CFAuthProxy struct {
	listener     net.Listener

	gatewayURL *url.URL
	addr       string

	authMiddleware func(http.Handler) http.Handler
}

type Oauth2ClientReader interface {
	Read(token string) (Oauth2Client, error)
}

func NewCFAuthProxy(gatewayAddr, addr string, oauth2Client Oauth2ClientReader, scopes ...string) *CFAuthProxy {
	gatewayURL, err := url.Parse(gatewayAddr)
	if err != nil {
		log.Fatalf("failed to parse gateway address: %s", err)
	}

	p := &CFAuthProxy{
		gatewayURL: gatewayURL,
		addr:       addr,
		authMiddleware: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authToken := r.Header.Get("Authorization")
				if authToken == "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				c, err := oauth2Client.Read(authToken)
				if err != nil {
					log.Printf("failed to read from Oauth2 server: %s", err)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if !hasRequiredScope(scopes, c.Scopes) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				h.ServeHTTP(w, r)
			})
		},
	}

	return p
}

func hasRequiredScope(requiredScopes, clientScopes []string) bool {
	for _, requiredScope := range requiredScopes {
		for _, clientScope := range clientScopes {
			if requiredScope == clientScope {
				return true
			}
		}
	}
	return false
}

// Start starts the HTTP listener and serves the HTTP server. If the
// CFAuthProxy was initialized with the WithCFAuthProxyBlock option this
// method will block.
func (p *CFAuthProxy) Start() {
	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		log.Fatalf("failed to start listener: %s", err)
	}

	p.listener = ln

	server := http.Server{
		Handler: p.authMiddleware(p.reverseProxy()),
	}

	log.Fatal(server.Serve(ln))
}

// Addr returns the listener address. This must be called after calling Start.
func (p *CFAuthProxy) Addr() string {
	return p.listener.Addr().String()
}

func (p *CFAuthProxy) reverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(p.gatewayURL)
}
