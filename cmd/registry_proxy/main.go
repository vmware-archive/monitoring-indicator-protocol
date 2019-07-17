package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"net/http"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/mtls"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry_proxy"
)

type addrs []string

func (a *addrs) String() string {
	return strings.Join(*a, ",")
}

func (a *addrs) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	port := flag.Int("port", 10567, "Port to expose registration endpoints")
	host := flag.String("host", "", "Host to bind to for registration endpoints")
	serverPEM := flag.String("tls-pem-path", "", "Server TLS public cert pem path")
	serverKey := flag.String("tls-key-path", "", "Server TLS private key path")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	clientPEM := flag.String("tls-client-pem-path", "", "Client TLS public cert pem path")
	clientKey := flag.String("tls-client-key-path", "", "Client TLS private key path")
	serverCommonName := flag.String("tls-server-cn", "localhost", "server (indicator registry) common name")
	var registryAddrs addrs
	flag.Var(&registryAddrs, "registry-addr", "Registry addr to proxy write requests to")
	localRegistryAddr := flag.String("local-registry-addr", "", "Registry addr that is local to this proxy")
	flag.Parse()

	address := fmt.Sprintf("%s:%d", *host, *port)

	tlsConfig, err := mtls.NewServerConfig(*rootCACert)
	if err != nil {
		log.Fatalf("Error with creating mTLS server config: %s", err)
	}
	tlsClientConfig, err := mtls.NewClientConfig(*clientPEM, *clientKey, *rootCACert, *serverCommonName)
	if err != nil {
		log.Fatalf("Error with creating mTLS client config: %s", err)
	}

	localRegistryHandler, registryHandlers, err := createHandlers(*localRegistryAddr, registryAddrs, tlsClientConfig)
	if err != nil {
		log.Fatalf("Error with creating handlers: %s", err)
	}

	server := &http.Server{
		Addr:         address,
		Handler:      registry_proxy.NewHandler(localRegistryHandler, registryHandlers),
		TLSConfig:    tlsConfig,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Listening for request on port %d", *port)
	_ = server.ListenAndServeTLS(*serverPEM, *serverKey)
	log.Fatalf("Listen unblocked")
}

func createHandlers(localRegistryAddr string, registryAddrs []string, tlsConfig *tls.Config) (http.Handler, []http.Handler, error) {
	// When speaking to the local registry, do so via regular, unencrypted, no auth HTTP.
	localRegistryURL, err := url.Parse("http://" + localRegistryAddr)
	if err != nil {
		return nil, nil, err
	}
	localRegistryHandler := httputil.NewSingleHostReverseProxy(localRegistryURL)

	localRegistryHandler.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	var registryHandlers []http.Handler
	for _, addr := range registryAddrs {
		// When speaking with remote registry proxies, do so via mutual auth TLS.
		addrURL, err := url.Parse("https://" + addr)
		if err != nil {
			return nil, nil, err
		}

		h := httputil.NewSingleHostReverseProxy(addrURL)
		// This transport was copied from:
		//    https://github.com/golang/go/blob/5ec14065dcc4c066ca7e434be7239c942f0c2e5b/src/net/http/transport.go#L42-L54
		// We copied it here in order to add our TLS config. We didn't want
		// to modify the global http.DefaultTransport
		h.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		registryHandlers = append(registryHandlers, h)
	}
	return localRegistryHandler, registryHandlers, nil
}
