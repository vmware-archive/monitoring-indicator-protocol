package grafana_test

import (
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/grafana"

	. "github.com/onsi/gomega"
)

func TestEmptyIndicatorDocument(t *testing.T) {
	g := NewGomegaWithT(t)

	var doc *v1.IndicatorDocument
	_, err := grafana.ConfigMap(doc, nil, v1.UndefinedType)
	g.Expect(err).To(HaveOccurred())
}

func TestNoLayoutGeneratesDefaultDashboard(t *testing.T) {
	g := NewGomegaWithT(t)

	doc := &v1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
			UID:       types.UID("test-uid"),
		},
		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{
				Name:    "my_app",
				Version: "1.0.1",
			},
			Indicators: []v1.IndicatorSpec{
				{
					Name:   "latency",
					PromQL: "histogram_quantile(0.9, latency)",
					Thresholds: []v1.Threshold{
						{
							Level:    "critical",
							Operator: v1.GreaterThanOrEqualTo,
							Value:    float64(100.2),
						},
					},
					Documentation: map[string]string{
						"title": "90th Percentile Latency",
					},
				},
			},
		},
	}

	cm, err := grafana.ConfigMap(doc, func(document v1.IndicatorDocument) ([]byte, error) {
		return []byte("the-expected-json"), nil
	}, v1.UndefinedType)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Name).To(Equal("indicator-protocol-grafana-dashboard.test-namespace.test-name"))
	g.Expect(cm.Data["indicator-protocol-grafana-dashboard.test-namespace.test-name.json"]).To(Equal("the-expected-json"))
	g.Expect(cm.Labels["grafana_dashboard"]).To(Equal("true"))
}

func TestEmptyDashboard(t *testing.T) {
	g := NewGomegaWithT(t)

	doc := &v1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
			UID:       types.UID("test-uid"),
		},
		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{
				Name:    "my_app",
				Version: "1.0.1",
			},
			Indicators: []v1.IndicatorSpec{
				{
					Name:   "latency",
					PromQL: "histogram_quantile(0.9, latency)",
					Type:   v1.ServiceLevelIndicator,
					Thresholds: []v1.Threshold{
						{
							Level:    "critical",
							Operator: v1.GreaterThanOrEqualTo,
							Value:    float64(100.2),
						},
					},
					Documentation: map[string]string{
						"title": "90th Percentile Latency",
					},
				},
			},
		},
	}

	returned, err := grafana.ConfigMap(doc, nil, v1.KeyPerformanceIndicator)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(returned).To(BeNil())
}

func TestSetsUpOwnership(t *testing.T) {
	g := NewGomegaWithT(t)

	uid := types.UID("test-uid")
	doc := &v1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
			UID:       uid,
		},
		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{
				Name:    "my_app",
				Version: "1.0.1",
			},
			Indicators: []v1.IndicatorSpec{
				{
					Name:   "latency",
					PromQL: "histogram_quantile(0.9, latency)",
					Thresholds: []v1.Threshold{
						{
							Level:    "critical",
							Operator: v1.GreaterThanOrEqualTo,
							Value:    float64(100.2),
						},
					},
					Documentation: map[string]string{
						"title": "90th Percentile Latency",
					},
				},
			},
		},
	}

	cm, err := grafana.ConfigMap(doc, func(document v1.IndicatorDocument) ([]byte, error) {
		return []byte("the-expected-json"), nil
	}, v1.UndefinedType)

	truePtr := true
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Labels["owner"]).To(Equal("test-name-test-namespace"))
	g.Expect(cm.OwnerReferences).To(ConsistOf(metav1.OwnerReference{
		APIVersion: api_versions.V1,
		Kind:       "IndicatorDocument",
		Name:       "test-name",
		UID:        uid,
		Controller: &truePtr,
	}))
}

func TestDashboardMapperError(t *testing.T) {
	g := NewGomegaWithT(t)

	doc := &v1.IndicatorDocument{}

	_, err := grafana.ConfigMap(doc, func(document v1.IndicatorDocument) ([]byte, error) {
		return nil, errors.New("some-error")
	}, v1.UndefinedType)

	g.Expect(err).To(HaveOccurred())
}
