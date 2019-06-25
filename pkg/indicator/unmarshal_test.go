package indicator_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
)

func TestDocumentFromYAML(t *testing.T) {
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
  serviceLevel:
    objective: 99

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

		indie := indicator.Indicator{
			Name:   "test_performance_indicator",
			PromQL: `prom{deployment="$deployment"}`,
			Thresholds: []indicator.Threshold{{
				Level:    "warning",
				Operator: indicator.LessThanOrEqualTo,
				Value:    500,
			}},
			Alert: indicator.Alert{
				For:  "1m",
				Step: "1m",
			},
			ServiceLevel: &indicator.ServiceLevel{
				Objective: float64(99),
			},
			Presentation: indicator.Presentation{
				CurrentValue: false,
				ChartType:    indicator.StepChart,
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
		g.Expect(doc).To(BeEquivalentTo(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Metadata:   map[string]string{"deployment": "well-performing-deployment"},
			Indicators: []indicator.Indicator{
				indie,
			},
			Layout: indicator.Layout{
				Title:       "Monitoring Test Product",
				Description: "Test description",
				Sections: []indicator.Section{{
					Title:       "Test Section",
					Description: "This section includes indicators and metrics",
					Indicators:  []string{indie.Name},
				}},
			},
		}))
	})

	t.Run("returns empty list of indicators", func(t *testing.T) {
		t.Run("bad document", func(t *testing.T) {
			g := NewGomegaWithT(t)

			reader := ioutil.NopCloser(strings.NewReader(`---
indicators: []`))
			d, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(d.Indicators).To(HaveLen(0))
		})
	})

	t.Run("returns error if YAML is bad", func(t *testing.T) {
		t.Run("bad document", func(t *testing.T) {
			g := NewGomegaWithT(t)

			reader := ioutil.NopCloser(strings.NewReader(`--`))
			_, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).To(HaveOccurred())
		})
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

			g.Expect(d.Indicators[0].Alert).To(Equal(indicator.Alert{
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

			g.Expect(d.Indicators[0].Alert).To(Equal(
				indicator.Alert{
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

			g.Expect(d.Indicators[0].Alert).To(Equal(indicator.Alert{
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

			g.Expect(d.Layout).To(Equal(indicator.Layout{
				Sections: []indicator.Section{{
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

			g.Expect(d.Indicators[0].Presentation).To(BeEquivalentTo(indicator.Presentation{
				ChartType:    "step",
				CurrentValue: false,
				Frequency:    0,
				Labels:       []string{},
			}))
			g.Expect(d.Indicators[1].Presentation).To(BeEquivalentTo(indicator.Presentation{
				ChartType:    "step",
				CurrentValue: true,
				Frequency:    0,
				Labels:       []string{},
			}))
		})

		t.Run("it sets a default service level with a value of nil if none is provided", func(t *testing.T) {
			g := NewGomegaWithT(t)
			reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: v0
product:
  name: test_product
  version: 0.0.1
indicators:
- name: test_performance_indicator
  promql: prom{deployment="$deployment"}
`))
			d, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(d.Indicators[0].ServiceLevel).To(BeNil())
		})
	})

	t.Run("handles thresholds", func(t *testing.T) {
		t.Run("it handles all the operators", func(t *testing.T) {
			g := NewGomegaWithT(t)

			reader := ioutil.NopCloser(strings.NewReader(`---
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

			g.Expect(d.Indicators[0].Thresholds).To(Equal([]indicator.Threshold{
				{
					Level:    "warning",
					Operator: indicator.LessThan,
					Value:    0,
				},
				{
					Level:    "warning",
					Operator: indicator.LessThanOrEqualTo,
					Value:    1.2,
				},
				{
					Level:    "warning",
					Operator: indicator.EqualTo,
					Value:    0.2,
				},
				{
					Level:    "warning",
					Operator: indicator.NotEqualTo,
					Value:    123,
				},
				{
					Level:    "warning",
					Operator: indicator.GreaterThanOrEqualTo,
					Value:    642,
				},
				{
					Level:    "warning",
					Operator: indicator.GreaterThan,
					Value:    1.222225,
				},
			}))
		})

		t.Run("it returns undefined operator if there is no value", func(t *testing.T) {
			g := NewGomegaWithT(t)

			reader := ioutil.NopCloser(strings.NewReader(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - level: warning
  `))

			d, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(d.Indicators[0].Thresholds[0].Operator).To(Equal(indicator.Undefined))
			g.Expect(d.Indicators[0].Thresholds[0].Value).To(Equal(float64(0)))
		})

		t.Run("it returns an error if value is not a number", func(t *testing.T) {
			g := NewGomegaWithT(t)

			reader := ioutil.NopCloser(strings.NewReader(`---
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

		t.Run("it returns an error if multiple operators are provided", func(t *testing.T) {
			// TODO: this test changes the behavior of the yaml parsing validation
			//       it replaces "it picks one operator when multiple are provided"
			t.Skip("not implemented")

			g := NewGomegaWithT(t)

			reader := ioutil.NopCloser(strings.NewReader(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - gte: 10
    lt: 20
    level: warning
  `))

			_, err := indicator.DocumentFromYAML(reader)
			g.Expect(err).To(HaveOccurred())
		})

		t.Run("it picks one operator when multiple are provided", func(t *testing.T) {
			g := NewGomegaWithT(t)

			reader := ioutil.NopCloser(strings.NewReader(`---
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
			g.Expect(doc.Indicators[0].Thresholds[0]).To(BeEquivalentTo(indicator.Threshold{
				Level:    "warning",
				Operator: indicator.LessThan,
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

			g.Expect(d.Indicators[0].Presentation.ChartType).To(Equal(indicator.StepChart))
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

			g.Expect(d.Indicators[0].Presentation.ChartType).To(Equal(indicator.BarChart))
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

			g.Expect(d.Indicators[0].Presentation.ChartType).To(Equal(indicator.StatusChart))
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

			g.Expect(d.Indicators[0].Presentation.ChartType).To(Equal(indicator.QuotaChart))
		})
	})
}

func TestPatchFromYAML(t *testing.T) {
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
  path: /indicators/0/thresholds?/-
  value:
    level: warning
    gt: 100
`))
		p, err := indicator.PatchFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		var patchedThreshold interface{}
		patchedThreshold = map[interface{}]interface{}{
			"level": "warning",
			"gt":    100,
		}
		expectedPatch := indicator.Patch{
			APIVersion: "v0/patch",
			Match: indicator.Match{
				Name:    test_fixtures.StrPtr("my-other-component"),
				Version: test_fixtures.StrPtr("1.2.3"),
			},
			Operations: []patch.OpDefinition{{
				Type:  "replace",
				Path:  test_fixtures.StrPtr("/indicators/0/thresholds?/-"),
				Value: &patchedThreshold,
			}},
		}

		g.Expect(p).To(BeEquivalentTo(expectedPatch))
	})

	t.Run("parses empty product name and version", func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
apiVersion: test-apiversion

match:
  metadata:
    deployment: test-deployment

operations:
- type: replace
  path: /indicators/name=success_percentage
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
}

func TestProductFromYAML(t *testing.T) {
	t.Run("parses all the fields", func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
product:
  name: indi-pro
  version: 1.2.3
`))
		p, err := indicator.ProductFromYAML(reader)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(p).To(BeEquivalentTo(indicator.Product{
			Name:    "indi-pro",
			Version: "1.2.3",
		}))
	})
}

func TestMetadataFromYAML(t *testing.T) {
	t.Run("parses all the fields", func(t *testing.T) {
		g := NewGomegaWithT(t)
		reader := ioutil.NopCloser(strings.NewReader(`---
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
	t.Run("does not mess up thresholds", func(t *testing.T) {
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
		g.Expect(resultDoc.Indicators[0].Thresholds[0]).To(BeEquivalentTo(indicator.Threshold{
			Level:    "critical",
			Operator: indicator.NotEqualTo,
			Value:    100,
		}))
	})
}
