package main

import (
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"crypto/tls"
	"crypto/x509"

	"flag"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log := log.New(os.Stderr, "", log.LstdFlags)
	log.Print("Starting Indiciator Registry CF Auth Reverse Proxy...")
	defer log.Print("Closing Indiciator Registry CF Auth Reverse Proxy.")

	flagSet := flag.NewFlagSet("uaa-scoped-proxy", flag.ErrorHandling(0))
	backendURL := flagSet.String("backend-url", "", "The indicator registry address")
	listenAddr := flagSet.String("listen-addr", ":8081", "The public port to listen on")
	uaaURL := flagSet.String("uaa-url", "", "UAA server host (e.g. https://uaa.my-pcf.com)")
	uaaCAPath := flagSet.String("uaa-ca-path", "", "File path to root CA cert for UAA")
	uaaClientID := flagSet.String("log-cache-client", "", "the UAA client which has access to log-cache (doppler.firehose or logs.admin scope)")
	uaaClientSecret := flagSet.String("log-cache-client-secret", "", "the client secret")
	insecure := flagSet.Bool("k", false, "skips ssl verification (insecure)")

	flagSet.Parse(os.Args[1:])

	uaaClient := NewUAAClient(
		*uaaURL,
		*uaaClientID,
		*uaaClientSecret,
		buildUAAClient(*uaaCAPath, *insecure),
	)

	proxy := NewCFAuthProxy(
		*backendURL,
		*listenAddr,
		uaaClient,
		"notifications.write",
	)
	proxy.Start()
}

func buildUAAClient(uaaCAPath string, skipCertVerify bool) *http.Client {
	return &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipCertVerify,
				MinVersion:         tls.VersionTLS12,
				RootCAs:            loadUaaCA(uaaCAPath),
			},
			DisableKeepAlives: true,
		},
	}
}

func loadUaaCA(uaaCertPath string) *x509.CertPool {
	caCert, err := ioutil.ReadFile(uaaCertPath)
	if err != nil {
		log.Fatalf("failed to read UAA CA certificate: %s", err)
	}

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatal("failed to parse UAA CA certificate.")
	}

	return certPool
}
