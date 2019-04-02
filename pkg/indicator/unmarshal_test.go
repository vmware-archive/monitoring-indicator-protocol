package indicator_test

import (
	"testing"
	"time"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
)

func TestReturnsCompleteDocument(t *testing.T) {
	t.Run("it can parse all document fields", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product: 
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: <%= spec.deployment %>

indicators:
- name: test_performance_indicator
  documentation:
    title: Test Performance Indicator
    description: This is a valid markdown description.
    recommendedResponse: Panic!
    thresholdNote: Threshold Note Text
  promql: prom{deployment="$deployment"}
  presentation:
    currentValue: false
    chartType: step
    frequency: 5s
    labels:
    - job
    - ip
    units: nanoseconds
  thresholds:
  - level: warning
    gte: 50

layout:
  title: Monitoring Test Product
  description: Test description
  sections:
  - title: Test Section
    description: This section includes indicators and metrics
    indicators:
    - test_performance_indicator
`), indicator.SkipMetadataInterpolation, indicator.OverrideMetadata(map[string]string{"deployment": "well-performing-deployment"}))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(BeEquivalentTo(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
			Metadata:   map[string]string{"deployment": "well-performing-deployment"},
			Indicators: []indicator.Indicator{
				{
					Name:   "test_performance_indicator",
					PromQL: `prom{deployment="$deployment"}`,
					Alert: indicator.Alert{
						For:  "1m",
						Step: "1m",
					},
					Thresholds: []indicator.Threshold{
						{
							Level:    "warning",
							Operator: indicator.GreaterThanOrEqualTo,
							Value:    50,
						},
					},
					Presentation: &indicator.Presentation{
						CurrentValue: false,
						ChartType:    indicator.StepChart,
						Frequency:    5 * time.Second,
						Labels:       []string{"job", "ip"},
						Units:        "nanoseconds",
					},
					Documentation: map[string]string{
						"title":               "Test Performance Indicator",
						"description":         "This is a valid markdown description.",
						"recommendedResponse": "Panic!",
						"thresholdNote":       "Threshold Note Text",
					},
				},
			},
			Layout: indicator.Layout{
				Title:       "Monitoring Test Product",
				Description: "Test description",
				Sections: []indicator.Section{{
					Title:       "Test Section",
					Description: "This section includes indicators and metrics",
					Indicators: []indicator.Indicator{{
						Name:   "test_performance_indicator",
						PromQL: `prom{deployment="$deployment"}`,
						Alert: indicator.Alert{
							For:  "1m",
							Step: "1m",
						},
						Thresholds: []indicator.Threshold{
							{
								Level:    "warning",
								Operator: indicator.GreaterThanOrEqualTo,
								Value:    50,
							},
						},
						Presentation: &indicator.Presentation{
							ChartType: indicator.StepChart,
							Frequency: 5 * time.Second,
							Labels:    []string{"job", "ip"},
							Units:     "nanoseconds",
						},
						Documentation: map[string]string{
							"title":               "Test Performance Indicator",
							"description":         "This is a valid markdown description.",
							"recommendedResponse": "Panic!",
							"thresholdNote":       "Threshold Note Text",
						},
					}},
				}},
			},
		}))
	})

	t.Run("it uses defaults in the case of empty presentation data", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product:
  name: test_product
  version: 0.0.1
metadata:
  deployment: test_deployment

indicators:
- name: test_performance_indicator
  promql: prom{deployment="$deployment"}

layout:
  sections:
  - title: Metrics
    indicators:
    - test_performance_indicator

`))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(*d.Indicators[0].Presentation).To(BeEquivalentTo(indicator.Presentation{
			ChartType:    "step",
			CurrentValue: false,
			Frequency:    0,
			Labels:       []string{},
		}))
	})

	t.Run("it sets chartType to 'step' if none is provided", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product:
  name: test_product
  version: 0.0.1
metadata:
  deployment: test_deployment

indicators:
- name: test_performance_indicator
  promql: prom{deployment="$deployment"}
  presentation:
    currentValue: false

layout:
  sections:
  - title: Metrics
    indicators:
    - test_performance_indicator

`))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Presentation.ChartType).To(BeEquivalentTo("step"))
	})
}

func TestReturnsAnEmptyListWhenNoIndicatorsArePassed(t *testing.T) {
	g := NewGomegaWithT(t)

	d, err := indicator.ReadIndicatorDocument([]byte(`---
indicators: []`))
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(d.Indicators).To(HaveLen(0))
}

func TestConvertsThresholdsProperly(t *testing.T) {
	g := NewGomegaWithT(t)

	d, err := indicator.ReadIndicatorDocument([]byte(`---
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
}

func TestReturnsAnErrorIfTheYAMLIsUnparsable(t *testing.T) {
	t.Run("bad document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := indicator.ReadIndicatorDocument([]byte(`--`))
		g.Expect(err).To(HaveOccurred())
	})
}

func TestReturnsUndefinedOperatorIfThresholdHasNoValue(t *testing.T) {
	g := NewGomegaWithT(t)

	d, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - level: warning
  `))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(d.Indicators[0].Thresholds[0].Operator).To(Equal(indicator.Undefined))
	g.Expect(d.Indicators[0].Thresholds[0].Value).To(Equal(float64(0)))
}

func TestReturnsAnErrorIfAThresholdHasABadFloatValue(t *testing.T) {
	g := NewGomegaWithT(t)

	_, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - gte: abs
    level: warning
  `))
	g.Expect(err).To(HaveOccurred())
}

func TestReturnsErrors(t *testing.T) {
	t.Run("if layout section references non-existent indicator", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product:
  name: my-product
  version: 1
indicators: []
layout:
  title: docs
  description: desc
  sections:
  - title: metric section
    description: metric desc
    indicators:
    - not_found
  `))
		g.Expect(err).To(MatchError(ContainSubstring("documentation.sections[0].indicators[0] references non-existent indicator")))
	})
}

func TestReturnsDefaultLayoutWhenGivenNoLayout(t *testing.T) {
	g := NewGomegaWithT(t)
	d, err := indicator.ReadIndicatorDocument([]byte(`---
apiVersion: v0
product:
  name: well-performing-component
  version: 0.0.1
metadata:
  deployment: valid-deployment

indicators:
- name: test_performance_indicator
  promql: promql_test_expr
`))
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(d.Layout).To(Equal(indicator.Layout{
		Sections: []indicator.Section{{
			Title: "Metrics",
			Indicators: []indicator.Indicator{
				{
					Name:   "test_performance_indicator",
					PromQL: "promql_test_expr",
					Alert: indicator.Alert{
						For:  "1m",
						Step: "1m",
					},
					Presentation: &indicator.Presentation{
						CurrentValue: false,
						ChartType:    "step",
						Frequency:    0,
						Labels:       []string{},
					},
				},
			},
		}},
	}))
}

func TestReturnsACompletePatchDocument(t *testing.T) {
	g := NewGomegaWithT(t)

	var val interface{}
	val = map[interface{}]interface{}{
		"promql": `success_percentage_promql{source_id="origin"}`,
		"documentation": map[interface{}]interface{}{
			"title": "Success Percentage",
		}}

	name := "my-component"
	version := "1.2.3"
	indicatorPatch := indicator.Patch{
		APIVersion: "test-apiversion",
		Match: indicator.Match{
			Name:    &name,
			Version: &version,
		},
		Operations: []patch.OpDefinition{{
			Type:  "replace",
			Path:  strPtr("/indicators/name=success_percentage"),
			Value: &val,
		}},
	}

	documentBytes := []byte(`---
apiVersion: test-apiversion

match:
  product:
    name: my-component
    version: 1.2.3

operations:
- type: replace
  path: /indicators/name=success_percentage
  value:
    promql: success_percentage_promql{source_id="origin"}
    documentation:
      title: Success Percentage

`)
	p, err := indicator.ReadPatchBytes(documentBytes)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(p).To(BeEquivalentTo(indicatorPatch))
}

func TestReturnsPatchDocumentWithBlankMatchNameAndVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	documentBytes := []byte(`---
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

`)
	p, err := indicator.ReadPatchBytes(documentBytes)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(p.Match.Name).To(BeNil())
	g.Expect(p.Match.Version).To(BeNil())
}

func TestDocumentMatching(t *testing.T) {
	name1 := "testing"
	version1 := "123"
	matcher1 := indicator.Match{
		Name:    &name1,
		Version: &version1,
	}

	matcher2 := indicator.Match{
		Name:    nil,
		Version: nil,
		Metadata: map[string]string{
			"deployment": "test-deployment",
		},
	}

	name2 := "other-testing"
	version2 := "456"
	matcher3 := indicator.Match{
		Name:    &name2,
		Version: &version2,
		Metadata: map[string]string{
			"deployment": "other-test-deployment",
		},
	}

	t.Run("name and version", func(t *testing.T) {
		g := NewGomegaWithT(t)

		documentBytes := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: non-matching-test-deployment

indicators:
- name: test_indicator
  promql: test_expr
`)

		g.Expect(indicator.MatchDocument(matcher1, documentBytes)).To(BeTrue())
		g.Expect(indicator.MatchDocument(matcher2, documentBytes)).To(BeFalse())
		g.Expect(indicator.MatchDocument(matcher3, documentBytes)).To(BeFalse())
	})

	t.Run("metadata", func(t *testing.T) {
		g := NewGomegaWithT(t)

		documentBytes := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing-foo-foo
  version: 123456

metadata:
  deployment: test-deployment

indicators:
- name: test_indicator
  promql: test_expr
`)

		g.Expect(indicator.MatchDocument(matcher1, documentBytes)).To(BeFalse())
		g.Expect(indicator.MatchDocument(matcher2, documentBytes)).To(BeTrue())
		g.Expect(indicator.MatchDocument(matcher3, documentBytes)).To(BeFalse())
	})

	t.Run("name and version and metadata", func(t *testing.T) {
		g := NewGomegaWithT(t)

		documentBytes := []byte(`---
apiVersion: test-apiversion/document

product:
  name: other-testing
  version: 456

metadata:
  deployment: other-test-deployment

indicators:
- name: test_indicator
  promql: test_expr
`)

		g.Expect(indicator.MatchDocument(matcher1, documentBytes)).To(BeFalse())
		g.Expect(indicator.MatchDocument(matcher2, documentBytes)).To(BeFalse())
		g.Expect(indicator.MatchDocument(matcher3, documentBytes)).To(BeTrue())
	})
}

func TestPatching(t *testing.T) {
	t.Run("patches files that match", func(t *testing.T) {
		g := NewGomegaWithT(t)

		matchingDocument := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: test-deployment

indicators:
- name: test_indicator
  promql: test_expr
`)
		var val interface{} = "patched_promql"
		indicatorPatch := []indicator.Patch{{
			APIVersion: "test-apiversion/patch",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "replace",
					Path:  strPtr("/indicators/0/promql"),
					Value: &val,
				},
			},
		}}

		patchedBytes, err := indicator.ApplyPatches(indicatorPatch, matchingDocument)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].PromQL).To(BeEquivalentTo("patched_promql"))
	})

	t.Run("does not patch files that do not match", func(t *testing.T) {
		g := NewGomegaWithT(t)

		nonMatchingDocument := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: not-test-deployment

indicators:
- name: test_indicator
  promql: test_expr
`)
		var val interface{} = "patched_promql"
		indicatorPatch := []indicator.Patch{{
			APIVersion: "test-apiversion/patch",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "replace",
					Path:  strPtr("/indicators/0/promql"),
					Value: &val,
				},
			},
		}}

		unpatchedBytes, err := indicator.ApplyPatches(indicatorPatch, nonMatchingDocument)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(unpatchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].PromQL).To(BeEquivalentTo("test_expr"))
	})

	t.Run("replaces by index", func(t *testing.T) {
		g := NewGomegaWithT(t)

		var patchedThreshold interface{} = map[interface{}]interface{}{
			"level": "warning",
			"gt":    "1000",
		}

		indicatorPatch := []indicator.Patch{{
			APIVersion: "test-apiversion/patch",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "replace",
					Path:  strPtr("/indicators/1/thresholds/1"),
					Value: &patchedThreshold,
				},
			},
		}}
		doc := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: test-deployment

indicators:
- name: test_indicator
  promql: test_expr
- name: test_indicator_2
  promql: test_expr
  thresholds: 
  - level: critical
    gt: 1500
  - level: warning
    gt: 500
`)

		patchedBytes, err := indicator.ApplyPatches(indicatorPatch, doc)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[1].Thresholds[1]).To(BeEquivalentTo(indicator.Threshold{
			Level:    "warning",
			Operator: indicator.GreaterThan,
			Value:    1000,
		}))
	})

	t.Run("replaces by attribute value", func(t *testing.T) {
		g := NewGomegaWithT(t)

		var patchedThreshold interface{} = map[interface{}]interface{}{
			"level": "warning",
			"gt":    "800",
		}

		indicatorPatch := []indicator.Patch{{
			APIVersion: "test-apiversion/patch",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "replace",
					Path:  strPtr("/indicators/name=test_indicator/thresholds/level=warning"),
					Value: &patchedThreshold,
				},
			},
		}}
		doc := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: test-deployment

indicators:
- name: test_indicator
  promql: test_expr
  thresholds:
  - level: warning
    gt: 500    
  - level: critical
    gt: 1000
`)

		patchedBytes, err := indicator.ApplyPatches(indicatorPatch, doc)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Thresholds[0]).To(BeEquivalentTo(indicator.Threshold{
			Level:    "warning",
			Operator: indicator.GreaterThan,
			Value:    800,
		}))
	})

	t.Run("removes", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicatorPatch := []indicator.Patch{{
			APIVersion: "test-apiversion/patch",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "remove",
					Path:  strPtr("/indicators/0/thresholds/level=warning"),
					Value: nil,
				},
			},
		}}
		doc := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: test-deployment

indicators:
- name: test_indicator
  promql: test_expr
  thresholds:
  - level: warning
    gt: 500
  - level: critical
    gt: 1000
`)

		patchedBytes, err := indicator.ApplyPatches(indicatorPatch, doc)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Thresholds).To(HaveLen(1))
	})

	t.Run("ignores `test` operation", func(t *testing.T) {
		g := NewGomegaWithT(t)

		var testVal interface{} = "not_test_indicator"
		indicatorPatch := []indicator.Patch{{
			APIVersion: "v0",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "test",
					Path:  strPtr("/indicators/0/name"),
					Value: &testVal,
				},
				{
					Type:  "remove",
					Path:  strPtr("/indicators/0/thresholds/level=warning"),
					Value: nil,
				},
			},
		}}
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
  - level: warning
    gt: 500
`)

		patchedBytes, err := indicator.ApplyPatches(indicatorPatch, doc)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Thresholds).To(HaveLen(0))
	})

	t.Run("adds by replacing", func(t *testing.T) {
		g := NewGomegaWithT(t)

		var newThresholds interface{} = map[interface{}]interface{}{
			"level": "warning",
			"gt":    "10",
		}

		indicatorPatch := []indicator.Patch{{
			APIVersion: "test-apiversion/patch",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "replace",
					Path:  strPtr("/indicators/name=test_indicator/thresholds?/-"),
					Value: &newThresholds,
				},
			},
		}}
		doc := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: test-deployment

indicators:
- name: test_indicator
  promql: test_expr
`)

		patchedBytes, err := indicator.ApplyPatches(indicatorPatch, doc)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Thresholds).To(HaveLen(1))
	})

	t.Run("does not error when patch fails due to invalid operation", func(t *testing.T) {
		g := NewGomegaWithT(t)

		indicatorPatch := []indicator.Patch{{
			APIVersion: "v0",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type: "replace",
					Path: strPtr("/indicators/name=test_indicator/thresholds?/-"),
				},
			},
		}}

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
`)

		patchedBytes, err := indicator.ProcessDocument(indicatorPatch, doc)
		g.Expect(err).To(BeEmpty())

		d, err2 := indicator.ReadIndicatorDocument(doc)
		g.Expect(patchedBytes).To(Equal(d))
		g.Expect(err2).ToNot(HaveOccurred())
	})

	t.Run("does not error when patch fails due to invalid path", func(t *testing.T) {
		g := NewGomegaWithT(t)

		var val interface{} = "patched_threshold"
		indicatorPatch := []indicator.Patch{{
			APIVersion: "v0",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "replace",
					Path:  strPtr("/indicators/35/thresholds/0"),
					Value: &val,
				},
			},
		}}
		//^ OpDefinition does not contain value

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
`)

		patchedBytes, err := indicator.ProcessDocument(indicatorPatch, doc)
		g.Expect(err).To(BeEmpty())

		d, err2 := indicator.ReadIndicatorDocument(doc)
		g.Expect(patchedBytes).To(Equal(d))
		g.Expect(err2).ToNot(HaveOccurred())
	})

	t.Run("applies partially successful patches", func(t *testing.T) {
		g := NewGomegaWithT(t)

		var patchedWarningThreshold interface{} = map[interface{}]interface{}{
			"level": "warning",
			"gt":    "800",
		}
		var patchedCriticalThreshold interface{} = map[interface{}]interface{}{
			"level": "critical",
			"gt":    "5000",
		}
		var patchedPromql interface{} = "foo"

		indicatorPatch := []indicator.Patch{{
			APIVersion: "test-apiversion/patch",
			Match: indicator.Match{
				Metadata: map[string]string{
					"deployment": "test-deployment",
				},
			},
			Operations: []patch.OpDefinition{
				{
					Type:  "replace",
					Path:  strPtr("/indicators/0/thresholds/level=warning"),
					Value: &patchedWarningThreshold,
				},
				{
					Type:  "replace",
					Path:  strPtr("/indicators/1/promql"),
					Value: &patchedPromql,
				},
				{
					Type:  "replace",
					Path:  strPtr("/indicators/0/thresholds/level=critical"),
					Value: &patchedCriticalThreshold,
				},
			},
		}}
		doc := []byte(`---
apiVersion: test-apiversion/document

product:
  name: testing
  version: 123

metadata:
  deployment: test-deployment

indicators:
- name: test_indicator
  promql: test_expr
  thresholds:
  - level: warning
    gt: 500    
  - level: critical
    gt: 1000
`)

		patchedBytes, err := indicator.ApplyPatches(indicatorPatch, doc)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].PromQL).To(Equal("test_expr"))
		g.Expect(d.Indicators[0].Thresholds).To(BeEquivalentTo([]indicator.Threshold{
			{
				Level:    "warning",
				Operator: indicator.GreaterThan,
				Value:    800,
			},
			{
				Level:    "critical",
				Operator: indicator.GreaterThan,
				Value:    5000,
			},
		}))
	})
}

func TestDefaultAlertConfig(t *testing.T) {
	t.Run("populates default alert config when no alert given", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument([]byte(`---
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
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Alert).To(Equal(indicator.Alert{
			For:  "1m",
			Step: "1m",
		}))
	})

	t.Run("populates default alert 'for' k/v when no alert for given", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument([]byte(`---
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
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Alert).To(Equal(
			indicator.Alert{
				For:  "1m",
				Step: "5m",
			}))
	})

	t.Run("populates default alert 'step' k/v when no alert step given", func(t *testing.T) {
		g := NewGomegaWithT(t)
		d, err := indicator.ReadIndicatorDocument([]byte(`---
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
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d.Indicators[0].Alert).To(Equal(indicator.Alert{
			For:  "5m",
			Step: "1m",
		}))
	})
}

func strPtr(s string) *string {
	return &s
}
