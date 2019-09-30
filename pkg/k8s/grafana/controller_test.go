package grafana_test

import (
	"errors"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/grafana"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestController(t *testing.T) {
	t.Run("it adds", func(t *testing.T) {
		g := NewGomegaWithT(t)
		spyConfigMapEditor := &spyConfigMapEditor{g: g}
		controller := grafana.NewController(spyConfigMapEditor, v1.UndefinedType)

		controller.OnAdd(indicatorDocument())

		spyConfigMapEditor.expectCreated([]*corev1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "indicator-protocol-grafana-dashboard.default.rabbit-mq-resource-name",
					Labels: map[string]string{
						"grafana_dashboard": "true",
						"owner":             "rabbit-mq-resource-name-default",
					},
				},
			},
		})
	})

	t.Run("on add it updates existing config map", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spyConfigMapEditor := &spyConfigMapEditor{g: g}
		spyConfigMapEditor.alreadyCreated()
		controller := grafana.NewController(spyConfigMapEditor, v1.UndefinedType)

		controller.OnAdd(indicatorDocument())

		spyConfigMapEditor.expectThatNothingWasCreated()
		spyConfigMapEditor.expectUpdated([]*corev1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "indicator-protocol-grafana-dashboard.default.rabbit-mq-resource-name",
					Labels: map[string]string{
						"grafana_dashboard": "true",
						"owner":             "rabbit-mq-resource-name-default",
					},
				},
			},
		})
	})

	t.Run("fails to add a non-indicator", func(t *testing.T) {
		g := NewGomegaWithT(t)
		spyConfigMapEditor := &spyConfigMapEditor{g: g}
		controller := grafana.NewController(spyConfigMapEditor, v1.UndefinedType)

		controller.OnAdd(666)

		spyConfigMapEditor.expectThatNothingWasCreated()
	})

	t.Run("it updates", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spyConfigMapEditor := &spyConfigMapEditor{g: g}
		p := grafana.NewController(spyConfigMapEditor, v1.UndefinedType)

		p.OnUpdate(nil, indicatorDocument())

		spyConfigMapEditor.expectUpdated([]*corev1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "indicator-protocol-grafana-dashboard.default.rabbit-mq-resource-name",
					Labels: map[string]string{
						"grafana_dashboard": "true",
						"owner":             "rabbit-mq-resource-name-default",
					},
				},
			},
		})
	})

	t.Run("fails to update a non-indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)
		spyConfigMapEditor := &spyConfigMapEditor{g: g}
		controller := grafana.NewController(spyConfigMapEditor, v1.UndefinedType)

		controller.OnUpdate(indicatorDocument(), 616)
		spyConfigMapEditor.expectThatNothingWasUpdated()

		controller.OnUpdate("asdf", indicatorDocument())
		spyConfigMapEditor.expectThatNothingWasUpdated()
	})

	t.Run("does not update when new and old objects are the same", func(t *testing.T) {
		g := NewGomegaWithT(t)
		spyConfigMapEditor := &spyConfigMapEditor{g: g}
		controller := grafana.NewController(spyConfigMapEditor, v1.UndefinedType)

		controller.OnUpdate(nil, nil)
		spyConfigMapEditor.expectThatNothingWasUpdated()
		controller.OnUpdate(indicatorDocument(), indicatorDocument())
		spyConfigMapEditor.expectThatNothingWasUpdated()
	})

	t.Run("only creates config maps for indicator type requested", func(t *testing.T) {
		g := NewGomegaWithT(t)
		spyConfigMapEditor := &spyConfigMapEditor{g: g}
		controller := grafana.NewController(spyConfigMapEditor, v1.ServiceLevelIndicator)

		indicators := []v1.IndicatorSpec{
			{
				Name:   "qps",
				PromQL: "rate(qps)",
				Type:   v1.KeyPerformanceIndicator,
			},
			{
				Name:   "lol",
				PromQL: "rate(fun)",
				Type:   v1.ServiceLevelIndicator,
			},
			{
				Name:   "foo",
				PromQL: "rate(boo)",
				Type:   v1.DefaultIndicator,
			},
		}
		controller.OnAdd(&v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rabbit-mq-resource-name",
				Namespace: "default",
				UID:       types.UID("some-uid"),
			},
			Spec: v1.IndicatorDocumentSpec{
				Product: v1.Product{
					Name:    "rabbit-mq-product-name",
					Version: "v1.0",
				},
				Indicators: indicators,
				Layout:     test_fixtures.DefaultLayout(indicators),
			},
		})

		createdConfigMap := spyConfigMapEditor.getCreated(0)
		g.Expect(createdConfigMap.Data["indicator-protocol-grafana-dashboard.default.rabbit-mq-resource-name.json"]).To(MatchJSON(`
{
  "title": "",
  "rows": [
	{
	  "title": "",
	  "panels": [
		{
		  "title": "lol",
		  "type": "graph",
		  "targets": [
			{
			  "expr": "rate(fun)"
			}
		  ],
		  "thresholds": null
		}
	  ]
	}
  ],
  "annotations": {
	"list": [
	  {
		"enable": true,
		"expr": "ALERTS{product=\"rabbit-mq-product-name\"}",
		"tagKeys": "level",
		"titleFormat": "{{alertname}} is {{alertstate}} in the {{level}} threshold",
		"iconColor": "#1f78c1"
	  }
	]
  }
}
`))
	})

	// TODO: test that a namespace provided to the controller is set in the cm objects
}

type spyConfigMapEditor struct {
	g *GomegaWithT

	getExists   bool
	createCalls []*corev1.ConfigMap
	updateCalls []*corev1.ConfigMap
}

func (s *spyConfigMapEditor) Create(cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	s.createCalls = append(s.createCalls, cm)
	return nil, nil
}

func (s *spyConfigMapEditor) Update(cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	s.updateCalls = append(s.updateCalls, cm)
	return nil, nil
}

func (s *spyConfigMapEditor) Get(name string, options metav1.GetOptions) (*corev1.ConfigMap, error) {
	if s.getExists {
		return nil, nil
	}
	return nil, errors.New("not found")
}

func (s *spyConfigMapEditor) alreadyCreated() {
	s.getExists = true
}

func (s *spyConfigMapEditor) expectCreated(cms []*corev1.ConfigMap) {
	s.g.Expect(s.createCalls).To(HaveLen(len(cms)))
	for i, cm := range cms {
		s.g.Expect(s.createCalls[i].Name).To(Equal(cm.Name))
		s.g.Expect(s.createCalls[i].Labels).To(Equal(cm.Labels))

		dashboardFilename := reflect.ValueOf(s.createCalls[i].Data).MapKeys()[0].String()
		s.g.Expect(s.createCalls[i].Data[dashboardFilename]).ToNot(BeEmpty())
	}
}

func (s *spyConfigMapEditor) expectUpdated(cms []*corev1.ConfigMap) {
	s.g.Expect(s.updateCalls).To(HaveLen(len(cms)))
	for i, cm := range cms {
		s.g.Expect(s.updateCalls[i].Name).To(Equal(cm.Name))
		s.g.Expect(s.updateCalls[i].Labels).To(Equal(cm.Labels))

		dashboardFilename := reflect.ValueOf(s.updateCalls[i].Data).MapKeys()[0].String()
		s.g.Expect(s.updateCalls[i].Data[dashboardFilename]).ToNot(BeEmpty())
	}
}

func (s *spyConfigMapEditor) expectThatNothingWasCreated() {
	s.g.Expect(s.createCalls).To(BeNil())
}
func (s *spyConfigMapEditor) expectThatNothingWasUpdated() {
	s.g.Expect(s.updateCalls).To(BeNil())
}

func (s *spyConfigMapEditor) getCreated(i int) *corev1.ConfigMap {
	return s.createCalls[i]
}

func indicatorDocument() *v1.IndicatorDocument {
	return &v1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbit-mq-resource-name",
			Namespace: "default",
			UID:       types.UID("some-uid"),
		},
		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{
				Name:    "rabbit-mq-product-name",
				Version: "v1.0",
			},
			Indicators: []v1.IndicatorSpec{
				{
					Name:   "qps",
					PromQL: "rate(qps)",
				},
			},
			Layout: v1.Layout{
				Title: "rabbit-mq-layout-title",
				Sections: []v1.Section{
					{
						Title:       "qps section",
						Description: "",
						Indicators:  []string{"qps"},
					},
				},
			},
		},
	}
}
