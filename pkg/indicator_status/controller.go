package indicator_status

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
	"github.com/prometheus/common/model"
)

type DocumentGetter interface {
	IndicatorDocuments() ([]registry.APIV0Document, error)
}
type StatusUpdater interface {
	BulkStatusUpdate(statusUpdates []registry.APIV0UpdateIndicatorStatus, documentId string) error
}

type PromQLClient interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, error)
}

type StatusController struct {
	documentGetter DocumentGetter
	statusUpdater  StatusUpdater
	interval       time.Duration
	promQLClient   PromQLClient
}

func NewStatusController(
	dg DocumentGetter,
	su StatusUpdater,
	pc PromQLClient,
	interval time.Duration,
) *StatusController {
	return &StatusController{
		documentGetter: dg,
		statusUpdater:  su,
		interval:       interval,
		promQLClient:   pc,
	}
}

func (c StatusController) Start() {
	err := c.updateStatuses()
	if err != nil {
		log.Printf("Failed to update indicator statuses: %s", err)
	}

	for {
		time.Sleep(c.interval)
		err := c.updateStatuses()
		if err != nil {
			log.Printf("Failed to update indicator statuses: %s", err)
		}
	}
}

func (c StatusController) updateStatuses() error {
	var statusUpdates []registry.APIV0UpdateIndicatorStatus

	apiv0Documents, err := c.documentGetter.IndicatorDocuments()
	if err != nil {
		return fmt.Errorf("error retrieving indicator docs: %s", err)
	}
	for _, indicatorDocument := range apiv0Documents {
		for _, indicator := range indicatorDocument.Indicators {
			if len(indicator.Thresholds) == 0 {
				continue
			}

			values, err := QueryValues(c.promQLClient, indicator.PromQL)
			if err != nil {
				log.Printf("Error querying Prometheus: %s", err)
				continue
			}
			thresholds := registry.ConvertThresholds(indicator.Thresholds)
			status := Match(thresholds, values)
			statusUpdates = append(statusUpdates, registry.APIV0UpdateIndicatorStatus{
				Name:   indicator.Name,
				Status: &status,
			})
		}
		err := c.statusUpdater.BulkStatusUpdate(statusUpdates, indicatorDocument.UID)
		if err != nil {
			log.Print(err)
		}
	}

	return nil
}
