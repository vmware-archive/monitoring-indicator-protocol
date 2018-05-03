package producer_test

import (
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/producer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Producer", func() {
	It("sends a log every specified period", func() {
		client := &fakeLoggregatorClient{}
		stop := producer.Start(client, 100*time.Millisecond)

		time.Sleep(300 * time.Millisecond)
		stop()

		Expect(client.GetCount()).To(BeNumerically(">=", 2))
		Expect(client.GetCounterName()).To(Equal("Eventproducerintervalcount"))
	})

	It("stops sending logs when the cleanup function is called", func() {
		client := &fakeLoggregatorClient{}
		stop := producer.Start(client, 100*time.Millisecond)

		time.Sleep(300 * time.Millisecond)
		stop()

		currentCount := client.GetCount()
		Consistently(client.GetCount).Should(Equal(currentCount))
	})
})

type fakeLoggregatorClient struct {
	counterName string
	count       int
	mu          sync.Mutex
}

func (o *fakeLoggregatorClient) EmitCounter(name string, opts ...loggregator.EmitCounterOption) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.counterName = name
	o.count++
}

func (o *fakeLoggregatorClient) GetCount() int {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.count
}

func (o *fakeLoggregatorClient) GetCounterName() string {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.counterName
}
