package grafana_test

import (
	"errors"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/grafana"
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
			Indicators: []v1alpha1.Indicator{
				{
					Name:   "latency",
					Promql: "histogram_quantile(0.9, latency)",
					Thresholds: []v1alpha1.Threshold{
						{
							Level: "critical",
							Gte:   floatVar(100.2),
						},
					},
					Documentation: map[string]string{
						"title": "90th Percentile Latency",
					},
				},
			},
		},
	}

	cm, err := grafana.ConfigMap(doc, func(indicator.Document) ([]byte, error) {
		return []byte("the-expected-json"), nil
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Name).To(Equal("test-name-77c8855f6"))
	g.Expect(cm.Data["my_app_3557e2a48e894b31ff24c4a09bf861a13001fa02.json"]).To(Equal("the-expected-json"))
	g.Expect(cm.Labels["grafana_dashboard"]).To(Equal("true"))
}

func TestDashboardMapperError(t *testing.T) {
	g := NewGomegaWithT(t)

	doc := &v1alpha1.IndicatorDocument{}

	_, err := grafana.ConfigMap(doc, func(indicator.Document) ([]byte, error) {
		return nil, errors.New("some-error")
	})

	g.Expect(err).To(HaveOccurred())
}

func floatVar(f float64) *float64 {
	return &f
}
