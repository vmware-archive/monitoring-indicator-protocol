package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
)

var supportedCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
}

func NewServer(address, serverPEM, serverKey, rootCACert string, handler http.Handler) (func() error, func() error, error) {
	caCert, err := ioutil.ReadFile(rootCACert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read root CA certificate: %s\n", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	server := &http.Server{
		Addr:    address,
		Handler: handler,
		TLSConfig: &tls.Config{
			ClientCAs:                caCertPool,
			ClientAuth:               tls.RequireAndVerifyClientCert,
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CipherSuites:             supportedCipherSuites,
		},
	}

	start := func() error { return server.ListenAndServeTLS(serverPEM, serverKey) }
	stop := func() error { return server.Close() }

	return start, stop, nil
}

func NewClient(clientCert, clientKey, rootCACert string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, err
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(rootCACert)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		RootCAs:                  caCertPool,
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: supportedCipherSuites,
	}
	tlsConfig.BuildNameToCertificate()
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}
