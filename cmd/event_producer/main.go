package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.cloudfoundry.org/go-log-cache"
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/cf-indicators/pkg/evaluator"
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"code.cloudfoundry.org/cf-indicators/pkg/producer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	loggregatorClient, closeClient := loggregatorClient()
	logCacheClient := logCacheClient()

	kpiBytes, err := ioutil.ReadFile(os.Getenv("KPI_FILE"))
	if err != nil {
		log.Fatalf("could not read kpis: %s", err)
	}

	d, err := indicator.ReadIndicatorDocument(kpiBytes)
	if err != nil {
		log.Fatalf("could not read kpis: %s", err)
	}

	finish := producer.Start(
		loggregatorClient,
		logCacheClient,
		evaluator.GetSatisfiedEvents,
		15*time.Second,
		d.Indicators,
	)

	waitForShutdownSignal()

	log.Println("Event Producer shutting down")
	finish()
	closeClient()
}

func loggregatorClient() (*loggregator.IngressClient, func()) {
	tlsConfig, err := loggregator.NewIngressTLSConfig(
		os.Getenv("AGENT_CA_FILE"),
		os.Getenv("AGENT_CERT_FILE"),
		os.Getenv("AGENT_KEY_FILE"),
	)
	if err != nil {
		log.Fatal("could not create TLS config", err)
	}

	client, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr("localhost:3458"),
		loggregator.WithTag("origin", "event_producer"),
	)

	return client, func() { client.CloseSend() }
}

func logCacheClient() *logcache.Client {
	creds, err := newTLSConfig(
		os.Getenv("LOG_CACHE_CA_FILE"),
		os.Getenv("LOG_CACHE_CERT_FILE"),
		os.Getenv("LOG_CACHE_KEY_FILE"),
		"log-cache",
	)
	if err != nil {
		log.Fatalf("failed to load TLS config for log-cache: %s", err)
	}

	client := logcache.NewClient(
		os.Getenv("LOG_CACHE_ADDR"),
		logcache.WithViaGRPC(
			grpc.WithTransportCredentials(credentials.NewTLS(creds)),
		),
	)

	return client
}

func waitForShutdownSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
}

func newTLSConfig(caPath, certPath, keyPath, cn string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		ServerName:         cn,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
	}

	caCertBytes, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		return nil, errors.New("cannot parse ca cert")
	}

	tlsConfig.RootCAs = caCertPool

	return tlsConfig, nil
}
