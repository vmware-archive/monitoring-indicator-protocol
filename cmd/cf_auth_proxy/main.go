package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/benbjohnson/clock"

	uaa "code.cloudfoundry.org/uaa-go-client"
	uaaConfig "code.cloudfoundry.org/uaa-go-client/config"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
)

func main() {
	log.Println("Running cf auth proxy")
	host := flag.String("host", "", "Host to bind to for registration endpoints")
	port := flag.String("port", "5000", "Port to bind to")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	clientPEM := flag.String("tls-client-pem-path", "", "Client TLS public cert pem path")
	clientKey := flag.String("tls-client-key-path", "", "Client TLS private key path")
	serverPEM := flag.String("tls-pem-path", "", "Server TLS public cert pem path")
	serverKey := flag.String("tls-key-path", "", "Server TLS private key path")
	uaaAddress := flag.String("uaa-addr", "", "Address of the UAA server against which to verify tokens")
	serverCommonName := flag.String("tls-server-cn", "localhost", "server (indicator registry) common name")
	registryUrlString := flag.String("registry-addr", "", "URL of the registry to proxy")
	flag.Parse()

	address := fmt.Sprintf("%s:%s", *host, *port)

	tlsClientConfig, err := mtls.NewClientConfig(*clientPEM, *clientKey, *rootCACert, *serverCommonName)
	if err != nil {
		log.Fatalf("Error with creating mTLS client config: %s", err)
	}

	targetURL, err := url.Parse(*registryUrlString)
	registryProxyHandler := httputil.NewSingleHostReverseProxy(targetURL)

	registryProxyHandler.Transport = &http.Transport{
		TLSClientConfig: tlsClientConfig,
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if err != nil {
		log.Fatalf("Error with creating handlers: %s", err)
	}

	var server = &http.Server{
		Addr:         address,
		Handler:      newUaaHandler(*uaaAddress, registryProxyHandler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("CF Auth Proxy listening for request on: https://%s\n", address)
	log.Fatalf("CF Auth proxy listen unblocked: %s", server.ListenAndServeTLS(*serverPEM, *serverKey))
}

// This http.Handler proxies to another handler, if and only if the incoming
// requests are GETs and they authenticate with either ?token=<token>
// or Authentication: Bearer <token>.
func newUaaHandler(uaaAddress string, registryProxyHandler http.Handler) *uaaValidatingHandler {

	isRequestFromAdmin := func(r *http.Request) bool {

		token := r.Header.Get("Authorization")
		// If not provided through the header, it's possible we're trying to
		// access this through a web browser with a ?token= param
		if token == "" {
			token = "bearer " + r.URL.Query().Get("token")
		}

		cfg := &uaaConfig.Config{
			UaaEndpoint:      uaaAddress,
			SkipVerification: true,
		}

		logger := lager.NewLogger("cf_auth_proxy main")
		uaaClient, err := uaa.NewClient(logger, cfg, clock.New())
		if err != nil {
			log.Fatal(err)
		}

		err = uaaClient.DecodeToken(token, "doppler.firehose", "logs.admin")
		if err != nil {
			log.Printf("error talking to uaa: %s", err)
			return false
		}
		return true
	}

	return &uaaValidatingHandler{
		handler:            registryProxyHandler,
		isRequestFromAdmin: isRequestFromAdmin,
	}
}

type uaaValidatingHandler struct {
	// The handler we forward to if the request is a GET and requestIsFromAdmin returns true
	handler http.Handler

	// Function that tells us whether the given request is from an administrator or not,
	// based on the credentials in the request.
	isRequestFromAdmin func(r *http.Request) bool
}

func (h *uaaValidatingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && h.isRequestFromAdmin(r) {
		h.handler.ServeHTTP(rw, r)
	} else {
		rw.WriteHeader(403)
	}
}