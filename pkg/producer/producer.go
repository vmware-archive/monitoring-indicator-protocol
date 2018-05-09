package producer

import (
	"context"
	"fmt"
	"log"
	"time"

	"code.cloudfoundry.org/go-log-cache"
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"code.cloudfoundry.org/go-loggregator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"
)

type loggregatorClient interface {
	EmitCounter(name string, opts ...loggregator.EmitCounterOption)
	EmitEvent(ctx context.Context, title, body string, opts ...loggregator.EmitEventOption) error
}

type logCacheClient interface {
	PromQL(ctx context.Context, query string, opts ...logcache.PromQLOption) (*logcache_v1.PromQL_QueryResult, error)
}

type eventGetter func(result *logcache_v1.PromQL_QueryResult, thresholds []kpi.Threshold) []kpi.Event

func Start(loggregatorClient loggregatorClient, logCacheClient logCacheClient, getSatisfiedEvents eventGetter, frequency time.Duration, kpis []kpi.KPI) (blockingCompleter func()) {
	ticker := time.NewTicker(frequency)
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Incrementing heartbeat counter to metron")
				loggregatorClient.EmitCounter("event_producer_evaluations_count")

				for _, k := range kpis {
					promQLResult, err := logCacheClient.PromQL(context.Background(), k.PromQL)
					if err != nil {
						fmt.Println("error when sending promql: ", err)
					}

					for _, event := range getSatisfiedEvents(promQLResult, k.Thresholds) {
						loggregatorClient.EmitEvent(
							context.Background(),
							k.Name,
							k.Description,
							loggregator.WithEnvelopeTags(event.Tags),
							loggregator.WithEnvelopeTag("value", fmt.Sprintf("%f", event.Value)),
							loggregator.WithEnvelopeTag("level", event.ThresholdLevel),
							loggregator.WithEnvelopeTag("threshold", fmt.Sprintf("%f", event.ThresholdValue)),
						)
					}
				}
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
