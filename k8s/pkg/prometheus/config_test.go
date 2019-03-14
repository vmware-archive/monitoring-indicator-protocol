package prometheus_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var indicators = []*v1alpha1.IndicatorDocument{
	{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my_app_indicators",
			Namespace: "monitoring",
			Labels: map[string]string{
				"environment": "staging",
			},
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
					Alert: v1alpha1.Alert{
						For:  "5m",
						Step: "10s",
					},
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
			Layout: v1alpha1.Layout{},
		},
	},
	{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my_production_app_indicators",
			Namespace: "monitoring",
			Labels: map[string]string{
				"environment": "production",
			},
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{
				Name:    "my_app",
				Version: "1.0.1",
			},
			Indicators: []v1alpha1.Indicator{
				{
					Name:   "average_latency",
					Promql: "average(latency)",
					Alert: v1alpha1.Alert{
						For:  "10m",
						Step: "10s",
					},
					Thresholds: []v1alpha1.Threshold{
						{
							Level: "warning",
							Neq:   floatVar(0),
						},
					},
					Documentation: map[string]string{
						"title": "Average Latency",
					},
				},
			},
			Layout: v1alpha1.Layout{},
		},
	},
}

func TestConfig(t *testing.T) {
	t.Run("it renders empty groups when there is no indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)
		p := prometheus.NewConfig()

		g.Expect(p.String()).To(MatchYAML("groups: []"))
	})

	t.Run("it renders a group for each indicator document", func(t *testing.T) {
		testCases := map[string]struct {
			Indicators []*v1alpha1.IndicatorDocument
			Expected   string
		}{
			"1 document": {
				Indicators: []*v1alpha1.IndicatorDocument{indicators[0]},
				Expected: `
                    groups:
                    - name: monitoring/my_app_indicators
                      rules:
                      - alert: latency
                        expr: histogram_quantile(0.9, latency) >= 100.2
                        for: 5m
                        labels:
                          product: my_app
                          version: 1.0.1
                          level: critical
                          environment: staging
                        annotations:
                          title: 90th Percentile Latency
                `,
			},
			"2 documents": {
				Indicators: indicators,
				Expected: `
                    groups:
                    - name: monitoring/my_app_indicators
                      rules:
                      - alert: latency
                        annotations:
                          title: 90th Percentile Latency
                        expr: histogram_quantile(0.9, latency) >= 100.2
                        for: 5m
                        labels:
                          environment: staging
                          level: critical
                          product: my_app
                          version: 1.0.1
                    - name: monitoring/my_production_app_indicators
                      rules:
                      - alert: average_latency
                        annotations:
                          title: Average Latency
                        expr: average(latency) != 0
                        for: 10m
                        labels:
                          environment: production
                          level: warning
                          product: my_app
                          version: 1.0.1
                `,
			},
		}

		for tn, tc := range testCases {
			t.Run(tn, func(t *testing.T) {
				g := NewGomegaWithT(t)
				p := prometheus.NewConfig()

				for _, i := range tc.Indicators {
					p.Upsert(i)
				}

				g.Expect(p.String()).To(MatchYAML(tc.Expected))
			})
		}
	})

	t.Run("it does not render deleted documents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		p := prometheus.NewConfig()

		for _, i := range indicators {
			p.Upsert(i)
		}

		p.Delete(indicators[0])

		g.Expect(p.String()).To(MatchYAML(`
            groups:
            - name: monitoring/my_production_app_indicators
              rules:
              - alert: average_latency
                annotations:
                  title: Average Latency
                expr: average(latency) != 0
                for: 10m
                labels:
                  environment: production
                  level: warning
                  product: my_app
                  version: 1.0.1
        `))
	})

	t.Run("it can handle concurrent reads/writes", func(t *testing.T) {
		p := prometheus.NewConfig()

		for i := 0; i < 100; i++ {
			go p.Upsert(indicators[0])
		}
		for i := 0; i < 100; i++ {
			go p.Delete(indicators[0])
		}
		for i := 0; i < 100; i++ {
			go p.String()
		}
	})
}

func floatVar(f float64) *float64 {
	return &f
}
