package grafana

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/domain"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mapper func(document indicator.Document) ([]byte, error)

var trueVal = true

func ConfigMap(doc *v1alpha1.IndicatorDocument, m mapper) (*v1.ConfigMap, error) {
	if doc == nil {
		return nil, errors.New("source indicator document was empty")
	}

	if m == nil {
		m = func(document indicator.Document) ([]byte, error) {
			grafanaDashboard, err := grafana_dashboard.DocumentToDashboard(document)
			if err != nil {
				return nil, err
			}
			dashboard := grafanaDashboard
			data, err := json.Marshal(dashboard)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}

	document := domain.Map(doc)
	jsonVal, err := m(document)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("indicator-protocol-grafana-dashboard.%s.%s", doc.ObjectMeta.Namespace, doc.ObjectMeta.Name)

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"grafana_dashboard": "true",
				"owner":             doc.Name + "-" + doc.Namespace,
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps.pivotal.io/v1alpha1",
				Kind:       "IndicatorDocument",
				Name:       doc.Name,
				UID:        doc.UID,
				Controller: &trueVal,
			}},
		},
		Data: map[string]string{
			fmt.Sprintf("%s.json", name): string(jsonVal),
		},
	}

	return cm, nil
}
