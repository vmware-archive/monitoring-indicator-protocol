package grafana

import (
	"encoding/json"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mapper func(document v1.IndicatorDocument) ([]byte, error)

var trueVal = true

func ConfigMap(doc *v1.IndicatorDocument, m mapper, indicatorType v1.IndicatorType) (*corev1.ConfigMap, error) {
	if doc == nil {
		return nil, errors.New("source indicator document was empty")
	}

	if m == nil {
		m = func(document v1.IndicatorDocument) ([]byte, error) {
			grafanaDashboard, err := grafana_dashboard.ToGrafanaDashboard(document, indicatorType)
			if err != nil {
				return nil, err
			}

			if grafanaDashboard == nil {
				return nil, nil
			}

			dashboard := grafanaDashboard
			data, err := json.Marshal(dashboard)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}

	jsonVal, err := m(*doc)
	if err != nil {
		return nil, err
	}

	if jsonVal == nil {
		return nil, nil
	}

	name := GenerateObjectName(doc)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"grafana_dashboard": "true",
				"owner":             doc.Name + "-" + doc.Namespace,
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: api_versions.V1,
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

func GenerateObjectName(doc *v1.IndicatorDocument) string {
	return fmt.Sprintf("indicator-protocol-grafana-dashboard.%s.%s", doc.ObjectMeta.Namespace, doc.ObjectMeta.Name)
}
