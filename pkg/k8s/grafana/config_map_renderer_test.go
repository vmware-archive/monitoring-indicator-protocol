package grafana_test

import (
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/grafana"

	. "github.com/onsi/gomega"
)

func TestEmptyIndicatorDocument(t *testing.T) {
	g := NewGomegaWithT(t)

	var doc *v1alpha1.IndicatorDocument
	_, err := grafana.ConfigMap(doc, nil)
	g.Expect(err).To(HaveOccurred())
}

func TestNoLayoutGeneratesDefaultDashboard(t *testing.T) {
	g := NewGomegaWithT(t)

	doc := &v1alpha1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
			UID:       types.UID("test-uid"),
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{
				Name:    "my_app",
				Version: "1.0.1",
			},
			Indicators: []v1alpha1.IndicatorSpec{
				{
					Name:   "latency",
					PromQL: "histogram_quantile(0.9, latency)",
					Thresholds: []v1alpha1.Threshold{
						{
							Level:    "critical",
							Operator: v1alpha1.GreaterThanOrEqualTo,
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

	cm, err := grafana.ConfigMap(doc, func(document v1alpha1.IndicatorDocument) ([]byte, error) {
		return []byte("the-expected-json"), nil
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Name).To(Equal("indicator-protocol-grafana-dashboard.test-namespace.test-name"))
	g.Expect(cm.Data["indicator-protocol-grafana-dashboard.test-namespace.test-name.json"]).To(Equal("the-expected-json"))
	g.Expect(cm.Labels["grafana_dashboard"]).To(Equal("true"))
}

func TestSetsUpOwnership(t *testing.T) {
	g := NewGomegaWithT(t)

	uid := types.UID("test-uid")
	doc := &v1alpha1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
			UID:       uid,
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{
				Name:    "my_app",
				Version: "1.0.1",
			},
			Indicators: []v1alpha1.IndicatorSpec{
				{
					Name:   "latency",
					PromQL: "histogram_quantile(0.9, latency)",
					Thresholds: []v1alpha1.Threshold{
						{
							Level:    "critical",
							Operator: v1alpha1.GreaterThanOrEqualTo,
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

	cm, err := grafana.ConfigMap(doc, func(document v1alpha1.IndicatorDocument) ([]byte, error) {
		return []byte("the-expected-json"), nil
	})

	truePtr := true
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Labels["owner"]).To(Equal("test-name-test-namespace"))
	g.Expect(cm.OwnerReferences).To(ConsistOf(metav1.OwnerReference{
		APIVersion: "apps.pivotal.io/v1alpha1",
		Kind:       "IndicatorDocument",
		Name:       "test-name",
		UID:        uid,
		Controller: &truePtr,
	}))
}

func TestDashboardMapperError(t *testing.T) {
	g := NewGomegaWithT(t)

	doc := &v1alpha1.IndicatorDocument{}

	_, err := grafana.ConfigMap(doc, func(document v1alpha1.IndicatorDocument) ([]byte, error) {
		return nil, errors.New("some-error")
	})

	g.Expect(err).To(HaveOccurred())
}
