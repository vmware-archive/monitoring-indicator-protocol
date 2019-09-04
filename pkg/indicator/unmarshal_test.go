package indicator_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestDocumentFromYAML(t *testing.T) {
	t.Run("returns error if YAML is bad", func(t *testing.T) {
		g := NewGomegaWithT(t)
		t.Run("bad document", func(t *testing.T) {
			reader := ioutil.NopCloser(strings.NewReader(`--`))
			_, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("apiVersion v0", func(t *testing.T) {
		t.Run("parses all document fields", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product: 
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: well-performing-deployment

indicators:
- name: test_performance_indicator
  documentation:
    title: Test Performance Indicator
    description: This is a valid markdown description.
    recommendedResponse: Panic!
    thresholdNote: Threshold Note Text
  thresholds:
  - level: warning
    lte: 500
  promql: prom{deployment="$deployment"}
  alert:
    for: 1m
    step: 1m
  presentation:
    currentValue: false
    chartType: step
    frequency: 5
    labels:
    - job
    - ip
    units: nanoseconds

layout:
  title: Monitoring Test Product
  description: Test description
  sections:
  - title: Test Section
    description: This section includes indicators and metrics
    indicators:
    - test_performance_indicator
`))
			doc, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			indie := v1.IndicatorSpec{
				Name:   "test_performance_indicator",
				PromQL: `prom{deployment="$deployment"}`,
				Thresholds: []v1.Threshold{{
					Level:    "warning",
					Operator: v1.LessThanOrEqualTo,
					Value:    500,
				}},
				Alert: v1.Alert{
					For:  "1m",
					Step: "1m",
				},
				Presentation: v1.Presentation{
					CurrentValue: false,
					ChartType:    v1.StepChart,
					Frequency:    5,
					Labels:       []string{"job", "ip"},
					Units:        "nanoseconds",
				},
				Documentation: map[string]string{
					"title":               "Test Performance Indicator",
					"description":         "This is a valid markdown description.",
					"recommendedResponse": "Panic!",
					"thresholdNote":       "Threshold Note Text",
				},
			}
			g.Expect(doc).To(BeEquivalentTo(v1.IndicatorDocument{
				TypeMeta: metav1.TypeMeta{
					APIVersion: api_versions.V0,
					Kind:       "IndicatorDocument",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"deployment": "well-performing-deployment"},
				},
				Spec: v1.IndicatorDocumentSpec{
					Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
					Indicators: []v1.IndicatorSpec{
						indie,
					},
					Layout: v1.Layout{
						Title:       "Monitoring Test Product",
						Description: "Test description",
						Sections: []v1.Section{{
							Title:       "Test Section",
							Description: "This section includes indicators and metrics",
							Indicators:  []string{indie.Name},
						}},
					},
				},
			}))
		})

		t.Run("populates defaults", func(t *testing.T) {
			t.Run("populates default alert config when no alert given", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: valid-deployment

indicators:
- name: test_indicator
  promql: promql_query
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Alert).To(Equal(v1.Alert{
					For:  "1m",
					Step: "1m",
				}))
			})

			t.Run("populates default alert 'for' k/v when no alert for given", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: valid-deployment

indicators:
- name: test_indicator
  promql: promql_query
  alert:
    step: 5m
`))

				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Alert).To(Equal(
					v1.Alert{
						For:  "1m",
						Step: "5m",
					}))
			})

			t.Run("populates default alert 'step' k/v when no alert step given", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: valid-deployment

indicators:
- name: test_indicator
  promql: promql_query
  alert:
    for: 5m
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Alert).To(Equal(v1.Alert{
					For:  "5m",
					Step: "1m",
				}))
			})

			t.Run("sets a default layout when not provided", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: valid-deployment

indicators:
- name: test_performance_indicator_1
  promql: promql_query
- name: test_performance_indicator_2
  promql: promql_query
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Layout).To(Equal(v1.Layout{
					Title: "well-performing-component - 0.0.1",
					Sections: []v1.Section{{
						Title: "Metrics",
						Indicators: []string{
							"test_performance_indicator_1",
							"test_performance_indicator_2",
						},
					}},
				}))
			})

			t.Run("it uses defaults in the case of empty presentation data", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: test_product
  version: 0.0.1
metadata:
  deployment: test_deployment

indicators:
- name: test_performance_indicator_1
  promql: prom{deployment="$deployment"}
- name: test_performance_indicator_2
  promql: prom{deployment="$deployment"}
  presentation:
    currentValue: true

layout:
  sections:
  - title: Metrics
    indicators:
    - test_performance_indicator

`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation).To(BeEquivalentTo(v1.Presentation{
					ChartType:    "step",
					CurrentValue: false,
					Frequency:    0,
					Labels:       []string{},
				}))
				g.Expect(d.Spec.Indicators[1].Presentation).To(BeEquivalentTo(v1.Presentation{
					ChartType:    "step",
					CurrentValue: true,
					Frequency:    0,
					Labels:       []string{},
				}))
			})

			t.Run("handles defaulting indicator types", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`---
apiVersion: v0
product:
  name: test_product
  version: 0.0.1

indicators:
- name: test_performance_indicator
  promql: prom{deployment="test"}
  presentation:
  chartType: step
`)))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(d.Spec.Indicators[0].Type).To(Equal(v1.DefaultIndicator))
			})
		})

		t.Run("handles thresholds", func(t *testing.T) {
			t.Run("it handles all the operators", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
indicators:
- name: test-kpi
  promql: prom
  thresholds:
  - lt: 0
    level: warning
  - lte: 1.2
    level: warning
  - eq: 0.2
    level: warning
  - neq: 123
    level: warning
  - gte: 642
    level: warning
  - gt: 1.222225
    level: warning`))

				d, err := indicator.DocumentFromYAML(reader)

				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Thresholds).To(Equal([]v1.Threshold{
					{
						Level:    "warning",
						Operator: v1.LessThan,
						Value:    0,
					},
					{
						Level:    "warning",
						Operator: v1.LessThanOrEqualTo,
						Value:    1.2,
					},
					{
						Level:    "warning",
						Operator: v1.EqualTo,
						Value:    0.2,
					},
					{
						Level:    "warning",
						Operator: v1.NotEqualTo,
						Value:    123,
					},
					{
						Level:    "warning",
						Operator: v1.GreaterThanOrEqualTo,
						Value:    642,
					},
					{
						Level:    "warning",
						Operator: v1.GreaterThan,
						Value:    1.222225,
					},
				}))
			})

			t.Run("it returns undefined operator if there is no value", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - level: warning
  `))

				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(d.Spec.Indicators[0].Thresholds[0].Operator).To(Equal(v1.UndefinedOperator))
				g.Expect(d.Spec.Indicators[0].Thresholds[0].Value).To(Equal(float64(0)))
			})

			t.Run("it returns an error if value is not a number", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - gte: abs
    level: warning
  `))

				_, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).To(HaveOccurred())
			})

			t.Run("it picks one operator when multiple are provided", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - gte: 10
    lt: 20
    level: warning
  `))

				doc, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(doc.Spec.Indicators[0].Thresholds[0]).To(BeEquivalentTo(v1.Threshold{
					Level:    "warning",
					Operator: v1.LessThan,
					Value:    20,
				}))
			})
		})

		t.Run("handles presentation chart types", func(t *testing.T) {
			t.Run("can set a step chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
 name: test_product
 version: 0.0.1

indicators:
- name: test_performance_indicator
  promql: prom{deployment="test"}
  presentation:
    chartType: step
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.StepChart))
			})

			t.Run("can set a bar chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
 name: test_product
 version: 0.0.1

indicators:
- name: test_performance_indicator
  promql: prom{deployment="test"}
  presentation:
    chartType: bar
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.BarChart))
			})

			t.Run("can set a status chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
 name: test_product
 version: 0.0.1

indicators:
- name: test_performance_indicator
  promql: prom{deployment="test"}
  presentation:
    chartType: status
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.StatusChart))
			})

			t.Run("can set a quota chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
 name: test_product
 version: 0.0.1
metadata:

indicators:
- name: test_performance_indicator
  promql: prom{deployment="test"}
  presentation:
    chartType: quota
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.QuotaChart))
			})
		})
	})

	t.Run("apiVersion v1", func(t *testing.T) {
		// TODO a lot of these tests don't need to work with YAML so extensively,
		//      they can be moved into tests for `PopulateDefaults`
		t.Run("parses all document fields", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  labels:
    deployment: well-performing-deployment

spec:
  product: 
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_performance_indicator
    documentation:
      title: Test Performance Indicator
      description: This is a valid markdown description.
      recommendedResponse: Panic!
      thresholdNote: Threshold Note Text
    thresholds:
    - level: warning
      operator: lte
      value: 500
    promql: prom{deployment="$deployment"}
    alert:
      for: 1m
      step: 1m
    presentation:
      currentValue: false
      chartType: step
      frequency: 5
      labels:
      - job
      - ip
      units: nanoseconds

  layout:
    title: Monitoring Test Product
    description: Test description
    sections:
    - title: Test Section
      description: This section includes indicators and metrics
      indicators:
      - test_performance_indicator
`))
			doc, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			indie := v1.IndicatorSpec{
				Name:   "test_performance_indicator",
				PromQL: `prom{deployment="$deployment"}`,
				Thresholds: []v1.Threshold{{
					Level:    "warning",
					Operator: v1.LessThanOrEqualTo,
					Value:    500,
				}},
				Alert: v1.Alert{
					For:  "1m",
					Step: "1m",
				},
				Presentation: v1.Presentation{
					CurrentValue: false,
					ChartType:    v1.StepChart,
					Frequency:    5,
					Labels:       []string{"job", "ip"},
					Units:        "nanoseconds",
				},
				Documentation: map[string]string{
					"title":               "Test Performance Indicator",
					"description":         "This is a valid markdown description.",
					"recommendedResponse": "Panic!",
					"thresholdNote":       "Threshold Note Text",
				},
			}
			g.Expect(doc).To(BeEquivalentTo(v1.IndicatorDocument{
				TypeMeta: metav1.TypeMeta{
					APIVersion: api_versions.V1,
					Kind: "IndicatorDocument",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"deployment": "well-performing-deployment"},
				},
				Spec: v1.IndicatorDocumentSpec{
					Product: v1.Product{Name: "well-performing-component", Version: "0.0.1"},
					Indicators: []v1.IndicatorSpec{
						indie,
					},
					Layout: v1.Layout{
						Title:       "Monitoring Test Product",
						Description: "Test description",
						Sections: []v1.Section{{
							Title:       "Test Section",
							Description: "This section includes indicators and metrics",
							Indicators:  []string{indie.Name},
						}},
					},
				},
			}))
		})

		t.Run("populates defaults", func(t *testing.T) {
			t.Run("populates default alert config when no alert given", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  labels:
    deployment: valid-deployment

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_indicator
    promql: promql_query
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Alert).To(Equal(v1.Alert{
					For:  "1m",
					Step: "1m",
				}))
			})

			t.Run("populates default alert 'for' k/v when no alert for given", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  labels:
    deployment: valid-deployment
spec:
  product:
    name: well-performing-component
    version: 0.0.1

  indicators:
  - name: test_indicator
    promql: promql_query
    alert:
      step: 5m
`))

				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Alert).To(Equal(
					v1.Alert{
						For:  "1m",
						Step: "5m",
					}))
			})

			t.Run("populates default alert 'step' k/v when no alert step given", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  labels:
    deployment: valid-deployment

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test_indicator
    promql: promql_query
    alert:
      for: 5m
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Alert).To(Equal(v1.Alert{
					For:  "5m",
					Step: "1m",
				}))
			})

			t.Run("sets a default layout when not provided", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  labels:
    deployment: valid-deployment

spec:
  product:
    name: well-performing-component
    version: 0.0.1

  indicators:
  - name: test_performance_indicator_1
    promql: promql_query
  - name: test_performance_indicator_2
    promql: promql_query
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Layout).To(Equal(v1.Layout{
					Title: "well-performing-component - 0.0.1",
					Sections: []v1.Section{{
						Title: "Metrics",
						Indicators: []string{
							"test_performance_indicator_1",
							"test_performance_indicator_2",
						},
					}},
				}))
			})

			t.Run("it uses defaults in the case of empty presentation data", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

metadata:
  labels:
    deployment: test_deployment

spec:
  product:
    name: test_product
    version: 0.0.1
  indicators:
  - name: test_performance_indicator_1
    promql: prom{deployment="$deployment"}
  - name: test_performance_indicator_2
    promql: prom{deployment="$deployment"}
    presentation:
      currentValue: true
  layout:
    sections:
    - title: Metrics
      indicators:
      - test_performance_indicator

`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation).To(BeEquivalentTo(v1.Presentation{
					ChartType:    "step",
					CurrentValue: false,
					Frequency:    0,
					Labels:       []string{},
				}))
				g.Expect(d.Spec.Indicators[1].Presentation).To(BeEquivalentTo(v1.Presentation{
					ChartType:    "step",
					CurrentValue: true,
					Frequency:    0,
					Labels:       []string{},
				}))
			})

			t.Run("handles defaulting indicator types", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument
spec:
  product:
    name: test_product
    version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: step
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(d.Spec.Indicators[0].Type).To(Equal(v1.DefaultIndicator))
			})

			t.Run("handles defaulting titles", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument
spec:
  product:
    name: test_product 
    version: 0.0.1
  layout:
    sections: []
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(d.Spec.Layout.Title).To(Equal("test_product - 0.0.1"))

			})
		})

		t.Run("handles thresholds", func(t *testing.T) {
			t.Run("it handles all the operators", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test-kpi
    promql: prom
    thresholds:
    - operator: lt
      value: 0
      level: warning
    - operator: lte
      value: 1.2
      level: warning
    - operator: eq
      value: 0.2
      level: warning
    - operator: neq
      value: 123
      level: warning
    - operator: gte
      value: 642
      level: warning
    - operator: gt
      value: 1.222225
      level: warning`))

				d, err := indicator.DocumentFromYAML(reader)

				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Thresholds).To(Equal([]v1.Threshold{
					{
						Level:    "warning",
						Operator: v1.LessThan,
						Value:    0,
					},
					{
						Level:    "warning",
						Operator: v1.LessThanOrEqualTo,
						Value:    1.2,
					},
					{
						Level:    "warning",
						Operator: v1.EqualTo,
						Value:    0.2,
					},
					{
						Level:    "warning",
						Operator: v1.NotEqualTo,
						Value:    123,
					},
					{
						Level:    "warning",
						Operator: v1.GreaterThanOrEqualTo,
						Value:    642,
					},
					{
						Level:    "warning",
						Operator: v1.GreaterThan,
						Value:    1.222225,
					},
				}))
			})

			t.Run("it handles unknown operator", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test-kpi
    description: desc
    promql: prom
    thresholds:
    - level: warning
      value: 500
      operator: foo
  `))

				_, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).To(HaveOccurred())
			})

			t.Run("it handles missing operator", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test-kpi
    description: desc
    promql: prom
    thresholds:
    - level: warning
  `))

				_, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).To(HaveOccurred())
			})

			t.Run("it returns an error if value is not a number", func(t *testing.T) {
				g := NewGomegaWithT(t)

				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
spec:
  product:
    name: well-performing-component
    version: 0.0.1
  indicators:
  - name: test-kpi
    description: desc
    promql: prom
    thresholds:
    - value: abs
      operator: gt
      level: warning
  `))

				_, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).To(HaveOccurred())
			})
		})

		t.Run("handles presentation chart types", func(t *testing.T) {
			t.Run("can set a step chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: step
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.StepChart))
			})

			t.Run("can set a bar chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: bar
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.BarChart))
			})

			t.Run("can set a status chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: status
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.StatusChart))
			})

			t.Run("can set a quota chartType", func(t *testing.T) {
				g := NewGomegaWithT(t)
				reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
    name: test_product
    version: 0.0.1

  indicators:
  - name: test_performance_indicator
    promql: prom{deployment="test"}
    presentation:
      chartType: quota
`))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(d.Spec.Indicators[0].Presentation.ChartType).To(Equal(v1.QuotaChart))
			})
		})

		t.Run("handles indicator types", func(t *testing.T) {
			getDoc := func(indiType string) string {
				return fmt.Sprintf(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument
spec:
  product:
   name: test_product
   version: 0.0.1

  indicators:
  - name: test_performance_indicator
    type: %s
    promql: prom{deployment="test"}
    presentation:
      chartType: step
`, indiType)
			}
			g := NewGomegaWithT(t)

			testCases := []struct {
				indiTypeString string
				indiType       v1.IndicatorType
			}{
				{"sli", v1.ServiceLevelIndicator},
				{"kpi", v1.KeyPerformanceIndicator},
				{"other", v1.DefaultIndicator},
			}

			for _, testCase := range testCases {
				yamlString := getDoc(testCase.indiTypeString)
				reader := ioutil.NopCloser(strings.NewReader(yamlString))
				d, err := indicator.DocumentFromYAML(reader)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(d.Spec.Indicators[0].Type).To(Equal(testCase.indiType),
					fmt.Sprintf("Failed indiTypeString: `%s`", testCase.indiTypeString))
			}

			yamlString := getDoc("")
			reader := ioutil.NopCloser(strings.NewReader(yamlString))
			_, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).To(HaveOccurred())
		})

	})

}

func TestPatchFromYAML(t *testing.T) {
	t.Run("apiVersion v0", func(t *testing.T) {
		t.Run("parses all the fields", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0/patch

match:
  product:
    name: my-other-component
    version: 1.2.3

operations:
- type: replace
  path: /spec/indicators/0/thresholds?/-
  value:
    level: warning
    operator: gt
    value: 100
`))
			p, err := indicator.PatchFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			var patchedThreshold interface{}
			patchedThreshold = map[string]interface{}{
				"level":    "warning",
				"operator": "gt",
				"value":    float64(100),
			}
			expectedPatch := indicator.Patch{
				APIVersion: api_versions.V0Patch,
				Match: indicator.Match{
					Name:    test_fixtures.StrPtr("my-other-component"),
					Version: test_fixtures.StrPtr("1.2.3"),
				},
				Operations: []patch.OpDefinition{{
					Type:  "replace",
					Path:  test_fixtures.StrPtr("/spec/indicators/0/thresholds?/-"),
					Value: &patchedThreshold,
				}},
			}

			g.Expect(p).To(BeEquivalentTo(expectedPatch))
		})

		t.Run("parses empty product name and version", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0/patch

match:
  metadata:
    deployment: test-deployment

operations:
- type: replace
  path: /spec/indicators/name=success_percentage
  value:
    promql: success_percentage_promql{source_id="origin"}
    documentation:
      title: Success Percentage

`))
			p, err := indicator.PatchFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(p.Match.Name).To(BeNil())
			g.Expect(p.Match.Version).To(BeNil())
		})
	})

	t.Run("apiVersion indicatorprotocol.io/v1", func(t *testing.T) {
		t.Run("parses all the fields", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocumentPatch

match:
  product:
    name: my-other-component
    version: 1.2.3

operations:
- type: replace
  path: /spec/indicators/0/thresholds?/-
  value:
    level: warning
    operator: gt
    value: 100
`))
			p, err := indicator.PatchFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			var patchedThreshold interface{}
			patchedThreshold = map[string]interface{}{
				"level":    "warning",
				"operator": "gt",
				"value":    float64(100),
			}
			expectedPatch := indicator.Patch{
				APIVersion: api_versions.V1,
				Match: indicator.Match{
					Name:    test_fixtures.StrPtr("my-other-component"),
					Version: test_fixtures.StrPtr("1.2.3"),
				},
				Operations: []patch.OpDefinition{{
					Type:  "replace",
					Path:  test_fixtures.StrPtr("/spec/indicators/0/thresholds?/-"),
					Value: &patchedThreshold,
				}},
			}

			g.Expect(p).To(BeEquivalentTo(expectedPatch))
		})

		t.Run("parses empty product name and version", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocumentPatch

match:
  metadata:
    deployment: test-deployment

operations:
- type: replace
  path: /spec/indicators/name=success_percentage
  value:
    promql: success_percentage_promql{source_id="origin"}
    documentation:
      title: Success Percentage

`))
			p, err := indicator.PatchFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(p.Match.Name).To(BeNil())
			g.Expect(p.Match.Version).To(BeNil())
		})
	})
}

func TestProductFromYAML(t *testing.T) {
	t.Run("api version v0", func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: indi-pro
  version: 1.2.3
`))
		p, err := indicator.ProductFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(p).To(BeEquivalentTo(v1.Product{
			Name:    "indi-pro",
			Version: "1.2.3",
		}))
	})

	t.Run(api_versions.V1, func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1
spec:
  product:
    name: indi-pro
    version: 1.2.3
`))
		p, err := indicator.ProductFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(p).To(BeEquivalentTo(v1.Product{
			Name:    "indi-pro",
			Version: "1.2.3",
		}))
	})
}

func TestMetadataFromYAML(t *testing.T) {
	t.Run("parses all the fields in v1 documents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: indicatorprotocol.io/v1

spec:
  product:
    name: indi-pro
    version: 1.2.3

metadata:
  labels:
    sound: meow
    size: small
    color: tabby
`))
		p, err := indicator.MetadataFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(p).To(BeEquivalentTo(map[string]string{
			"sound": "meow",
			"size":  "small",
			"color": "tabby",
		}))
	})

	t.Run("parses all the fields in v0 documents", func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0

product:
  name: indi-pro
  version: 1.2.3

metadata:
  sound: meow
  size: small
  color: tabby
`))
		p, err := indicator.MetadataFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(p).To(BeEquivalentTo(map[string]string{
			"sound": "meow",
			"size":  "small",
			"color": "tabby",
		}))
	})
}

func TestProcessesDocument(t *testing.T) {
	t.Run("does not mess up thresholds in apiVersion v0", func(t *testing.T) {
		g := NewGomegaWithT(t)
		doc := []byte(`---
apiVersion: v0
product:
  name: testing
  version: 123
metadata:
  deployment: test-deployment
indicators:
- name: test_indicator
  promql: test_expr
  thresholds:
  - level: critical
    neq: 100
`)
		resultDoc, err := indicator.ProcessDocument([]indicator.Patch{}, doc)
		g.Expect(err).To(HaveLen(0))
		g.Expect(resultDoc.Spec.Indicators[0].Thresholds[0]).To(BeEquivalentTo(v1.Threshold{
			Level:    "critical",
			Operator: v1.NotEqualTo,
			Value:    100,
		}))
	})

	t.Run("does not mess up thresholds in apiVersion v1", func(t *testing.T) {
		g := NewGomegaWithT(t)
		doc := []byte(`---
apiVersion: indicatorprotocol.io/v1
kind: IndicatorDocument

spec:
  product:
    name: testing
    version: v123
  indicators:
  - name: test_indicator
    promql: test_expr
    thresholds:
    - level: critical
      operator: neq
      value: 100
`)
		resultDoc, err := indicator.ProcessDocument([]indicator.Patch{}, doc)
		g.Expect(err).To(HaveLen(0))
		g.Expect(resultDoc.Spec.Indicators[0].Thresholds[0]).To(BeEquivalentTo(v1.Threshold{
			Level:    "critical",
			Operator: v1.NotEqualTo,
			Value:    100,
		}))
	})

}
