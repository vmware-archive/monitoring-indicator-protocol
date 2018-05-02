package producer

import (
	"log"
	"time"

	"code.cloudfoundry.org/go-loggregator"
)

type loggregatorClient interface {
	EmitCounter(name string, opts ...loggregator.EmitCounterOption)
}

func Start(client loggregatorClient, frequency time.Duration) (blockingCompleter func()) {
	ticker := time.NewTicker(frequency)
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Incrementing heartbeat counter to metron")
				client.EmitCounter("some-counter")
			case <-stop:
				return
			}
		}
	}()

	return func() {
		ticker.Stop()
		stop <- struct{}{}
	}
}
