package indicator_status

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/prometheus/common/model"
)

type registryClient interface {
	IndicatorDocuments() ([]registry.APIV0Document, error)
	BulkStatusUpdate(statusUpdates []registry.APIV0UpdateIndicatorStatus, documentId string) error
}

type promQLClient interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

type StatusController struct {
	RegistryClient registryClient
	IntervalTime   time.Duration
	PromQLClient   promQLClient
}

func (c StatusController) Start() {
	err := c.updateStatuses()
	if err != nil {
		log.Printf("failed to update indicator statuses: %s", err)
	}

	interval := time.NewTicker(c.IntervalTime)
	for {
		select {
		case <-interval.C:
			err := c.updateStatuses()
			if err != nil {
				log.Printf("failed to update indicator statuses: %s", err)
			}
		}
	}
}

func (c StatusController) updateStatuses() error {
	var statusUpdates []registry.APIV0UpdateIndicatorStatus

	apiv0Documents, err := c.RegistryClient.IndicatorDocuments()
	if err != nil {
		return fmt.Errorf("error retrieving indicator docs: %s", err)
	}
	for _, indicatorDocument := range apiv0Documents {
		for _, indicator := range indicatorDocument.Indicators {

			if len(indicator.Thresholds) == 0 {
				continue
			}

			value, err := c.PromQLClient.Query(context.Background(), indicator.PromQL, time.Time{})
			if err != nil {
				log.Printf("error querying Prometheus: %s", err)
			}

			vectors, ok := value.(model.Vector)
			if !ok {
				continue
			}

			var values []float64
			for _, v := range vectors {
				values = append(values, float64(v.Value))
			}
			status := Match(indicator.Thresholds, values)
			statusUpdates = append(statusUpdates, registry.APIV0UpdateIndicatorStatus{
				Name:   indicator.Name,
				Status: status,
			})
		}
		err := c.RegistryClient.BulkStatusUpdate(statusUpdates, indicatorDocument.UID)
		if err != nil {
			log.Printf("%s", err)
		}
	}

	return nil
}
