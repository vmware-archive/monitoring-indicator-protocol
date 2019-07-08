package indicator_status

import (
	"log"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

type DocumentGetter interface {
	IndicatorDocuments() ([]registry.APIV0Document, error)
}
type StatusUpdater interface {
	BulkStatusUpdate(statusUpdates []registry.APIV0UpdateIndicatorStatus, documentId string) error
}

type PromQLClient interface {
	QueryVectorValues(promql string) ([]float64, error)
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
		log.Print("Failed to update indicator statuses")
	}

	for {
		time.Sleep(c.interval)
		err := c.updateStatuses()
		if err != nil {
			log.Print("Failed to update indicator statuses")
		}
	}
}

func (c StatusController) updateStatuses() error {
	var statusUpdates []registry.APIV0UpdateIndicatorStatus

	apiv0Documents, err := c.documentGetter.IndicatorDocuments()
	if err != nil {
		return err
	}
	for _, indicatorDocument := range apiv0Documents {
		for _, indicator := range indicatorDocument.Indicators {
			if len(indicator.Thresholds) == 0 {
				continue
			}

			values, err := c.promQLClient.QueryVectorValues(indicator.PromQL)
			if err != nil {
				log.Print("Error querying Prometheus")
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
