package indicator_test

import (
	. "github.com/onsi/gomega"
	"testing"
	"time"

	"github.com/krishicks/yaml-patch"
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
						Frequency:    time.Duration(5 * time.Second),
						Labels:       []string{"job", "ip"},
					},
					Documentation: map[string]string{
						"title":                "Test Performance Indicator",
						"description":          "This is a valid markdown description.",
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
						Thresholds: []indicator.Threshold{
							{
								Level:    "warning",
								Operator: indicator.GreaterThanOrEqualTo,
								Value:    50,
							},
						},
						Presentation: &indicator.Presentation{
							ChartType: indicator.StepChart,
							Frequency: time.Duration(5 * time.Second),
							Labels:    []string{"job", "ip"},
						},
						Documentation: map[string]string{
							"title":                "Test Performance Indicator",
							"description":          "This is a valid markdown description.",
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

		g.Expect(d).To(BeEquivalentTo(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "test_product", Version: "0.0.1"},
			Metadata:   map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{
				{
					Name:   "test_performance_indicator",
					PromQL: `prom{deployment="test_deployment"}`,
					Presentation: &indicator.Presentation{
						ChartType:    "step",
						CurrentValue: false,
						Frequency:    0,
						Labels:       []string{},
					},
				},
			},
			Layout: indicator.Layout{
				Sections: []indicator.Section{{
					Title: "Metrics",
					Indicators: []indicator.Indicator{{
						Name:   "test_performance_indicator",
						PromQL: `prom{deployment="test_deployment"}`,
						Presentation: &indicator.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
					}},
				}},
			},
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

		g.Expect(d).To(BeEquivalentTo(indicator.Document{
			APIVersion: "v0",
			Product:    indicator.Product{Name: "test_product", Version: "0.0.1"},
			Metadata:   map[string]string{"deployment": "test_deployment"},
			Indicators: []indicator.Indicator{
				{
					Name:   "test_performance_indicator",
					PromQL: `prom{deployment="test_deployment"}`,
					Presentation: &indicator.Presentation{
						CurrentValue: false,
						ChartType:    "step",
						Frequency:    0,
						Labels:       []string{},
					},
				},
			},
			Layout: indicator.Layout{
				Sections: []indicator.Section{{
					Title: "Metrics",
					Indicators: []indicator.Indicator{{
						Name:   "test_performance_indicator",
						PromQL: `prom{deployment="test_deployment"}`,
						Presentation: &indicator.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
					}},
				}},
			},
		}))
	})
}

func TestReturnsAnEmptyListWhenNoIndicatorsArePassed(t *testing.T) {
	g := NewGomegaWithT(t)

	d, err := indicator.ReadIndicatorDocument([]byte(`---
indicators: []`))
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(d.Indicators).To(HaveLen(0))
}

func TestReturnsAConvertedIndicator(t *testing.T) {
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

	g.Expect(d.Indicators).To(Equal([]indicator.Indicator{{
		Name:   "test-kpi",
		PromQL: "prom",
		Thresholds: []indicator.Threshold{
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
		},
		Presentation: &indicator.Presentation{
			CurrentValue: false,
			ChartType:    "step",
			Frequency:    0,
			Labels: []string{},
		},
	}}))
}

func TestReturnsAnErrorIfTheYAMLIsUnparsable(t *testing.T) {
	t.Run("bad document", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := indicator.ReadIndicatorDocument([]byte(`--`))
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("bad chart type", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := indicator.ReadIndicatorDocument([]byte(`---
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
    chartType: bad-fake-no-good-chart`))

		g.Expect(err).To(MatchError(ContainSubstring("'bad-fake-no-good-chart' - valid chart types are step, bar")))
	})
}

func TestReturnsAnErrorIfAThresholdHasNoValue(t *testing.T) {
	g := NewGomegaWithT(t)

	_, err := indicator.ReadIndicatorDocument([]byte(`---
indicators:
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - level: warning
  `))
	g.Expect(err).To(HaveOccurred())
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
	t.Run("if section references non-existent indicator", func(t *testing.T) {
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

	g.Expect(d).To(Equal(indicator.Document{
		APIVersion: "v0",
		Product:    indicator.Product{Name: "well-performing-component", Version: "0.0.1"},
		Metadata:   map[string]string{"deployment": "valid-deployment"},
		Indicators: []indicator.Indicator{
			{
				Name:   "test_performance_indicator",
				PromQL: "promql_test_expr",
				Presentation: &indicator.Presentation{
					CurrentValue: false,
					ChartType:    "step",
					Frequency:    0,
					Labels:       []string{},
				},
			},
		},
		Layout: indicator.Layout{
			Sections: []indicator.Section{{
				Title: "Metrics",
				Indicators: []indicator.Indicator{
					{
						Name:   "test_performance_indicator",
						PromQL: "promql_test_expr",
						Presentation: &indicator.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels:       []string{},
						},
					},
				},
			}},
		},
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
	patch := indicator.Patch{
		APIVersion: "test-apiversion",
		Match: indicator.Match{
			Name:    &name,
			Version: &version,
		},
		Operations: []yamlpatch.Operation{{
			Op:    "add",
			Path:  "/indicators/name=success_percentage",
			Value: yamlpatch.NewNode(&val),
		}},
	}

	documentBytes := []byte(`---
apiVersion: test-apiversion

match:
  product:
    name: my-component
    version: 1.2.3

operations:
- op: add
  path: /indicators/name=success_percentage
  value:
    promql: success_percentage_promql{source_id="origin"}
    documentation:
      title: Success Percentage

`)
	p, err := indicator.ReadPatchBytes(documentBytes)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(p).To(BeEquivalentTo(patch))
}

func TestReturnsPatchDocumentWithBlankVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	var val interface{}
	val = map[interface{}]interface{}{
		"promql": `success_percentage_promql{source_id="origin"}`,
		"documentation": map[interface{}]interface{}{
			"title": "Success Percentage",
		}}

	patch := indicator.Patch{
		APIVersion: "test-apiversion",
		Match: indicator.Match{
			Name:    nil,
			Version: nil,
			Metadata: map[string]string{
				"deployment": "test-deployment",
			},
		},
		Operations: []yamlpatch.Operation{{
			Op:    "add",
			Path:  "/indicators/name=success_percentage",
			Value: yamlpatch.NewNode(&val),
		}},
	}

	documentBytes := []byte(`---
apiVersion: test-apiversion

match:
  metadata:
    deployment: test-deployment

operations:
- op: add
  path: /indicators/name=success_percentage
  value:
    promql: success_percentage_promql{source_id="origin"}
    documentation:
      title: Success Percentage

`)
	p, err := indicator.ReadPatchBytes(documentBytes)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(p).To(BeEquivalentTo(patch))
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
	var val interface{}
	val = map[interface{}]interface{}{
		"name":   "inserted_indicator",
		"promql": `inserted_indicator_promql{source_id="origin"}`,
		"documentation": map[interface{}]interface{}{
			"title": "Success Percentage",
		}}

	patch := []indicator.Patch{{
		APIVersion: "test-apiversion/patch",
		Match: indicator.Match{
			Metadata: map[string]string{
				"deployment": "test-deployment",
			},
		},
		Operations: []yamlpatch.Operation{{
			Op:    "add",
			Path:  "/indicators/-",
			Value: yamlpatch.NewNode(&val),
		}},
	}}

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

	nonMatchingDocument := []byte(`---
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

	t.Run("patches files that match", func(t *testing.T) {
		g := NewGomegaWithT(t)

		patchedBytes, err := indicator.ApplyPatches(patch, matchingDocument)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(patchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(BeEquivalentTo(indicator.Document{
			APIVersion: "test-apiversion/document",
			Product: indicator.Product{
				Name:    "testing",
				Version: "123",
			},
			Metadata: map[string]string{
				"deployment": "test-deployment",
			},
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: "test_expr",
				Presentation: &indicator.Presentation{
					CurrentValue: false,
					ChartType:    "step",
					Frequency:    0,
					Labels: []string{},
				},
			}, {
				Name:   "inserted_indicator",
				PromQL: `inserted_indicator_promql{source_id="origin"}`,
				Presentation: &indicator.Presentation{
					CurrentValue: false,
					ChartType:    "step",
					Frequency:    0,
					Labels: []string{},
				},
				Documentation: map[string]string{"title": "Success Percentage"},
			}},
			Layout: indicator.Layout{
				Sections: []indicator.Section{{
					Title:       "Metrics",
					Description: "",
					Indicators: []indicator.Indicator{{
						Name:   "test_indicator",
						PromQL: "test_expr",
						Presentation: &indicator.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels: []string{},
						},
					}, {
						Name:   "inserted_indicator",
						PromQL: `inserted_indicator_promql{source_id="origin"}`,
						Presentation: &indicator.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels: []string{},
						},
						Documentation: map[string]string{"title": "Success Percentage"},
					}},
				}},
			},
		}))
	})

	t.Run("does not patch files that do not match", func(t *testing.T) {
		g := NewGomegaWithT(t)
		unpatchedBytes, err := indicator.ApplyPatches(patch, nonMatchingDocument)
		g.Expect(err).ToNot(HaveOccurred())

		d, err := indicator.ReadIndicatorDocument(unpatchedBytes)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(d).To(BeEquivalentTo(indicator.Document{
			APIVersion: "test-apiversion/document",
			Product: indicator.Product{
				Name:    "testing",
				Version: "123",
			},
			Metadata: map[string]string{
				"deployment": "non-matching-test-deployment",
			},
			Indicators: []indicator.Indicator{{
				Name:   "test_indicator",
				PromQL: "test_expr",
				Presentation: &indicator.Presentation{
					CurrentValue: false,
					ChartType:    "step",
					Frequency:    0,
					Labels: []string{},
				},
			}},
			Layout: indicator.Layout{
				Sections: []indicator.Section{{
					Title:       "Metrics",
					Description: "",
					Indicators: []indicator.Indicator{{
						Name:   "test_indicator",
						PromQL: "test_expr",
						Presentation: &indicator.Presentation{
							CurrentValue: false,
							ChartType:    "step",
							Frequency:    0,
							Labels: []string{},
						},
					}},
				}},
			},
		}))
	})
}
