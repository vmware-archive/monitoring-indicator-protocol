package producer_test

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/go-log-cache"
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"
	"github.com/cloudfoundry-incubator/event-producer/pkg/producer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Producer", func() {
	It("evaluates KPIS on an interval and sends a heartbeat counter", func() {
		loggregatorClient := &fakeLoggregatorClient{}
		logCacheClient := &fakeLogCacheClient{}

		kpis := []kpi.KPI{{
			Name:        "latency",
			Description: "the latency metric",
			PromQL:      `latency{source_id="gorouter"}`,
			Thresholds: []kpi.Threshold{{
				Level:    "critical",
				Operator: kpi.GreaterThanOrEqualTo,
				Value:    1000,
			}},
		}}

		promQLResult := &logcache_v1.PromQL_QueryResult{
			Result: &logcache_v1.PromQL_QueryResult_Vector{
				Vector: &logcache_v1.PromQL_Vector{
					Samples: []*logcache_v1.PromQL_Sample{
						{
							Point: &logcache_v1.PromQL_Point{
								Time:  1525793488000000000,
								Value: 1001,
							},
						},
					},
				},
			},
		}

		eventGetter := func(result *logcache_v1.PromQL_QueryResult, thresholds []kpi.Threshold) []kpi.Event {
			Expect(result).To(Equal(promQLResult))

			return []kpi.Event{{
				Tags:           map[string]string{"ip": "127.0.0.1"},
				Value:          1001,
				ThresholdLevel: "critical",
				ThresholdValue: 1000,
			}}
		}

		logCacheClient.PromQLMethod = func(ctx context.Context, query string, opts ...logcache.PromQLOption) (*logcache_v1.PromQL_QueryResult, error) {
			Expect(query).To(Equal(`latency{source_id="gorouter"}`))

			return promQLResult, nil
		}

		stop := producer.Start(loggregatorClient, logCacheClient, eventGetter, 100*time.Millisecond, kpis)

		time.Sleep(300 * time.Millisecond)
		stop()

		Expect(loggregatorClient.GetEvents()).To(ContainElement(&loggregator_v2.Envelope{
			Timestamp:      0,
			SourceId:       "",
			InstanceId:     "",
			DeprecatedTags: nil,
			Tags: map[string]string{
				"ip":        "127.0.0.1",
				"value":     "1001.000000",
				"level":     "critical",
				"threshold": "1000.000000",
			},
			Message: &loggregator_v2.Envelope_Event{
				Event: &loggregator_v2.Event{
					Title: "latency",
					Body:  "the latency metric",
				},
			},
		}))

		Expect(loggregatorClient.GetCounterCount()).To(BeNumerically(">=", 2))
		Expect(loggregatorClient.GetCounterName()).To(Equal("event_producer_evaluations_count"))
	})

	It("stops sending logs when the cleanup function is called", func() {
		client := &fakeLoggregatorClient{}
		stop := producer.Start(client, nil, nil, 100*time.Millisecond, make([]kpi.KPI, 0))

		time.Sleep(300 * time.Millisecond)
		stop()

		currentCount := client.GetCounterCount()
		Consistently(client.GetCounterCount).Should(Equal(currentCount))
	})
})

type fakeLoggregatorClient struct {
	Err error

	counterName string
	counters    int
	events      []*loggregator_v2.Envelope
	mu          sync.Mutex
}

func (o *fakeLoggregatorClient) EmitCounter(name string, opts ...loggregator.EmitCounterOption) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.counterName = name
	o.counters++
}

func (o *fakeLoggregatorClient) EmitEvent(ctx context.Context, title, body string, opts ...loggregator.EmitEventOption) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.events == nil {
		o.events = make([]*loggregator_v2.Envelope, 0)
	}

	o.counters++
	message := loggregator_v2.Envelope{
		Tags: make(map[string]string),
		Message: &loggregator_v2.Envelope_Event{
			Event: &loggregator_v2.Event{
				Title: title,
				Body:  body,
			},
		},
	}
	for _, opt := range opts {
		opt(&message)
	}

	o.events = append(o.events, &message)

	return o.Err
}

func (o *fakeLoggregatorClient) GetEvents() []*loggregator_v2.Envelope {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.events
}

func (o *fakeLoggregatorClient) GetCounterCount() int {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.counters
}

func (o *fakeLoggregatorClient) GetCounterName() string {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.counterName
}

type fakeLogCacheClient struct {
	PromQLMethod func(ctx context.Context, query string, opts ...logcache.PromQLOption) (*logcache_v1.PromQL_QueryResult, error)
}

func (o *fakeLogCacheClient) PromQL(ctx context.Context, query string, opts ...logcache.PromQLOption) (*logcache_v1.PromQL_QueryResult, error) {
	return o.PromQLMethod(ctx, query, opts...)
}
