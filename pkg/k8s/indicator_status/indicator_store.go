package indicator_status

import (
	"sync"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

type IndicatorStore struct {
	sync.Mutex
	indicators []v1.Indicator
}

func NewIndicatorStore() *IndicatorStore {
	return &IndicatorStore{indicators: make([]v1.Indicator, 0)}
}

func (is *IndicatorStore) Add(indicator v1.Indicator) {
	is.Lock()
	is.indicators = append(is.indicators, indicator)
	is.Unlock()
}

func (is *IndicatorStore) Delete(indicator v1.Indicator) {
	nextIndicators := make([]v1.Indicator, 0)
	is.Lock()
	defer is.Unlock()
	for _, indie := range is.indicators {
		if !(indicator.Name == indie.Name) {
			nextIndicators = append(nextIndicators, indie)
		}
	}
	is.indicators = nextIndicators

}

func (is *IndicatorStore) Update(indicator v1.Indicator) {
	is.Lock()
	defer is.Unlock()
	for i, indie := range is.indicators {
		if indicator.Name == indie.Name {
			is.indicators[i] = indicator
		}
	}
}

func (is *IndicatorStore) GetIndicators() []v1.Indicator {
	is.Lock()
	indicators := make([]v1.Indicator, 0, len(is.indicators))
	for _, i := range is.indicators {
		indicators = append(indicators, i)
	}
	is.Unlock()
	return indicators
}
