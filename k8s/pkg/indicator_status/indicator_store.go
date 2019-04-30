package indicator_status

import (
	"sync"

	types "github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
)

type IndicatorStore struct {
	sync.Mutex
	indicators []types.Indicator
}

func NewIndicatorStore() IndicatorStore {
	return IndicatorStore{indicators: make([]types.Indicator, 0)}
}

func (is *IndicatorStore) Add(indicator types.Indicator) {
	is.Lock()
	is.indicators = append(is.indicators, indicator)
	is.Unlock()
}

func (is *IndicatorStore) Delete(indicator types.Indicator) {
	nextIndicators := make([]types.Indicator, 0)
	is.Lock()
	defer is.Unlock()
	for _, indie := range is.indicators {
		if !(indicator.Name == indie.Name) {
			nextIndicators = append(nextIndicators, indie)
		}
	}
	is.indicators = nextIndicators

}

func (is *IndicatorStore) Update(indicator types.Indicator) {
	is.Lock()
	defer is.Unlock()
	for i, indie := range is.indicators {
		if indicator.Name == indie.Name {
			is.indicators[i] = indicator
		}
	}
}

func (is *IndicatorStore) GetIndicators() []types.Indicator {
	is.Lock()
	indicators := make([]types.Indicator, 0, len(is.indicators))
	for _, i := range is.indicators {
		indicators = append(indicators, i)
	}
	is.Unlock()
	return indicators
}
