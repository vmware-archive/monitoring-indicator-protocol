package status_store

import (
	"fmt"
	"sync"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Clock func() time.Time

type UpdateRequest struct {
	Status        *string
	IndicatorName string
	DocumentUID   string
}

type IndicatorStatus struct {
	DocumentUID   string
	IndicatorName string
	Status        *string
	UpdatedAt     time.Time
}

func New(clock Clock) *Store {
	return &Store{clock: clock}
}

type Store struct {
	sync.Mutex
	statuses []IndicatorStatus
	clock    Clock
}

func (s *Store) UpdateStatus(request UpdateRequest) {
	s.Lock()
	defer s.Unlock()

	newStatus := IndicatorStatus{
		DocumentUID:   request.DocumentUID,
		IndicatorName: request.IndicatorName,
		Status:        request.Status,
		UpdatedAt:     s.clock(),
	}

	for idx, status := range s.statuses {
		if status.DocumentUID == request.DocumentUID && status.IndicatorName == request.IndicatorName {
			s.statuses[idx] = newStatus
			return
		}
	}

	s.statuses = append(s.statuses, newStatus)
}

func (s *Store) StatusFor(documentUID string, indicatorName string) (IndicatorStatus, error) {
	s.Lock()
	defer s.Unlock()

	for _, status := range s.statuses {
		if status.DocumentUID == documentUID && status.IndicatorName == indicatorName {
			return status, nil
		}
	}

	return IndicatorStatus{}, fmt.Errorf("indicator status for document %s with name %s could not be found", documentUID, indicatorName)
}

func (s *Store) FillStatuses(doc *v1alpha1.IndicatorDocument) {
	s.Lock()
	defer s.Unlock()

	docStatus := make(map[string]v1alpha1.IndicatorStatus)

	for _, status := range s.statuses {
		if status.DocumentUID == doc.BoshUID() {
			var newStatus v1alpha1.IndicatorStatus
			if status.Status != nil {
				newStatus.Phase = *status.Status
				newStatus.UpdatedAt = v1.Time{Time: status.UpdatedAt}
			}
			docStatus[status.IndicatorName] = newStatus
		}
	}

	doc.Status = docStatus
}
