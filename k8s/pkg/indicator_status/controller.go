package indicator_status

import (
	"log"
	"time"

	"github.com/benbjohnson/clock"
	types "github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/clientset/versioned/typed/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/domain"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator_status"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PromQLClient interface {
	QueryVectorValues(query string) ([]float64, error)
}

type indicatorStore interface {
	Add(indicator types.Indicator)
	Delete(indicator types.Indicator)
	Update(indicator types.Indicator)
	GetIndicators() []types.Indicator
}

type Controller struct {
	interval        time.Duration
	promqlClient    PromQLClient
	indicatorClient v1alpha1.IndicatorsGetter
	clock           clock.Clock
	namespace       string
	indicatorStore  indicatorStore
}

func NewController(
	indicatorClient v1alpha1.IndicatorsGetter,
	promqlClient PromQLClient,
	interval time.Duration,
	clock clock.Clock,
	namespace string,
	store indicatorStore,
) *Controller {
	return &Controller{
		interval:        interval,
		indicatorClient: indicatorClient,
		promqlClient:    promqlClient,
		indicatorStore:  store,
		clock:           clock,
		namespace:       namespace,
	}
}

func (c *Controller) Start() {
	existingList, err := c.indicatorClient.Indicators(c.namespace).List(v1.ListOptions{})
	if err != nil {
		log.Print("Could not load existing indicators on Start")
	}
	if existingList.Items != nil {
		for _, indicator := range existingList.Items {
			c.indicatorStore.Add(indicator)
		}
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
		log.Print("Invalid resource type OnAdd")
		return
	}

	c.indicatorStore.Add(*indicator)
}

func (c *Controller) OnUpdate(oldObj, newObj interface{}) {
	indicator, ok := newObj.(*types.Indicator)
	if !ok {
		log.Print("Invalid resource type OnUpdate")
		return
	}

	c.indicatorStore.Update(*indicator)
}

func (c *Controller) OnDelete(obj interface{}) {
	indicator, ok := obj.(*types.Indicator)
	if !ok {
		log.Print("Invalid resource type OnDelete")
		return
	}

	c.indicatorStore.Delete(*indicator)
}

func (c *Controller) updateStatuses() {
	for _, indicator := range c.indicatorStore.GetIndicators() {
		status, err := c.getStatus(indicator)
		if err != nil {
			log.Print("Error getting status for indicator")
			continue
		}

		if indicator.Status.Phase == status {
			continue
		}

		indicator.Status = types.IndicatorStatus{
			Phase: status,
		}
		_, err = c.indicatorClient.Indicators(indicator.Namespace).Update(&indicator)

		if err != nil {
			log.Print("Error updating indicator")
		}
	}
}

func (c *Controller) getStatus(indicator types.Indicator) (string, error) {
	status := indicator_status.Undefined
	if len(indicator.Spec.Thresholds) == 0 {
		return status, nil
	}
	values, err := c.promqlClient.QueryVectorValues(indicator.Spec.Promql)
	if err != nil {
		log.Print("Error querying Prometheus")
		return "", err
	}

	domainThresholds := domain.MapToDomainThreshold(indicator.Spec.Thresholds)
	status = indicator_status.Match(domainThresholds, values)
	return status, nil
}
