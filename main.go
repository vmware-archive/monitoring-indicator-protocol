package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.cloudfoundry.org/go-loggregator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/producer"
)

func main() {
	client, closeClient := loggregatorClient()

	finish := producer.Start(client, 15*time.Second)

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
		log.Fatal("Could not create TLS config", err)
	}

	client, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr("localhost:3458"),
		loggregator.WithTag("origin", "event_producer"),
	)

	return client, func() { client.CloseSend() }
}

func waitForShutdownSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
}
