package indicator_status

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	types "github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/clientset/versioned/typed/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/domain"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	"github.com/prometheus/common/model"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PromQLClient interface {
	// TODO: can we change this to queryValues?
	Query(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

type indicatorStore struct {
	sync.Mutex
	indicators []types.Indicator
}

type Controller struct {
	interval        time.Duration
	indicatorStore  *indicatorStore
	promqlClient    PromQLClient
	indicatorClient v1alpha1.IndicatorsGetter
	clock           clock.Clock
	namespace       string
}

func NewController(
	indicatorClient v1alpha1.IndicatorsGetter,
	promqlClient PromQLClient,
	interval time.Duration,
	clock clock.Clock,
	namespace string,
) *Controller {
	return &Controller{
		interval:        interval,
		indicatorClient: indicatorClient,
		promqlClient:    promqlClient,
		indicatorStore: &indicatorStore{
			indicators: make([]types.Indicator, 0),
		},
		clock:     clock,
		namespace: namespace,
	}
}

func (c *Controller) Start() {
	existingList, err := c.indicatorClient.Indicators(c.namespace).List(v1.ListOptions{})
	if err != nil {
		log.Printf("Could not load existing indicators on Start: %s", err)
	}
	if existingList.Items != nil {
		c.indicatorStore.Lock()
		c.indicatorStore.indicators = existingList.Items
		c.indicatorStore.Unlock()
	}
	c.updateStatuses()
	for {
		c.clock.Sleep(c.interval)
		c.updateStatuses()
	}
}

func (c *Controller) OnAdd(obj interface{}) {
	indicator, ok := obj.(*types.Indicator)
	if !ok {
		log.Printf("Invalid resource type OnAdd: %T", obj)
		return
	}
	c.indicatorStore.Lock()
	defer c.indicatorStore.Unlock()
	c.indicatorStore.indicators = append(c.indicatorStore.indicators, *indicator)
}

func (c *Controller) OnUpdate(oldObj, newObj interface{}) {
	indicator, ok := newObj.(*types.Indicator)
	if !ok {
		log.Printf("Invalid resource type OnUpdate: %T", indicator)
		return
	}

	c.indicatorStore.Lock()
	defer c.indicatorStore.Unlock()
	for i, indie := range c.indicatorStore.indicators {
		if indicator.Name == indie.Name {
			c.indicatorStore.indicators[i] = *indicator
		}
	}
}

func (c *Controller) OnDelete(obj interface{}) {
	indicator, ok := obj.(*types.Indicator)
	if !ok {
		log.Printf("Invalid resource type OnDelete: %T", obj)
		return
	}

	c.indicatorStore.Lock()
	defer c.indicatorStore.Unlock()
	nextIndicators := make([]types.Indicator, 0)
	for _, indie := range c.indicatorStore.indicators {
		if !(indicator.Name == indie.Name) {
			nextIndicators = append(nextIndicators, indie)
		}
	}
	c.indicatorStore.indicators = nextIndicators
}

func (c *Controller) updateStatuses() {
	//TODO: move status off of spec onto a subresource
	c.indicatorStore.Lock()
	defer c.indicatorStore.Unlock()

	for _, indicator := range c.indicatorStore.indicators {
		status, err := c.getStatus(indicator)
		if err != nil {
			log.Printf("Error getting status: %s", err)
			continue
		}

		updatedIndicator := indicator.DeepCopy()
		updatedIndicator.Spec.Status = status
		// TODO: [optimization] do not update status if no change
		_, err = c.indicatorClient.Indicators(indicator.Namespace).Update(updatedIndicator)
		if err != nil {
			log.Printf("Error updating indicator %s: %s", updatedIndicator.Name, err)
			continue
		}
	}
}

func (c *Controller) getStatus(indicator types.Indicator) (*string, error) {
	status := "UNDEFINED"
	if len(indicator.Spec.Thresholds) == 0 {
		return &status, nil
	}
	values, err := indicator_status.QueryValues(c.promqlClient, indicator.Spec.Promql)
	if err != nil {
		log.Printf("Error querying Prometheus: %s", err)
		return nil, err
	}

	domainThresholds := domain.MapToDomainThreshold(indicator.Spec.Thresholds)
	status = indicator_status.Match(domainThresholds, values)
	return &status, nil
}
