package mtls

// TODO rename to TLS, SingleAuthClient is not mutual TLS
import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
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

func NewServerConfig(caPath string) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, errors.New("failed to read root CA certificate file")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		ClientCAs:                caCertPool,
		ClientAuth:               tls.RequireAndVerifyClientCert,
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites:             supportedCipherSuites,
	}, nil
}

func NewClientConfig(clientCert, clientKey, rootCACert, serverCommonName string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, err
	}

	caCert, err := ioutil.ReadFile(rootCACert)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates:             []tls.Certificate{cert},
		RootCAs:                  caCertPool,
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites:             supportedCipherSuites,
		ServerName:               serverCommonName,
	}, nil
}

func NewSingleAuthClientConfig(rootCACert, serverCommonName string) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(rootCACert)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		RootCAs:                  caCertPool,
		CipherSuites:             supportedCipherSuites,
		ServerName:               serverCommonName,
	}, nil
}
