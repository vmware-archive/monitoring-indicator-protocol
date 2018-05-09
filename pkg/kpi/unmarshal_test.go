package kpi_test

import (
	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unmarshal", func() {
	It("returns an empty list when no kpis are passed", func() {
		kpis, err := kpi.ReadKPIsFromYaml([]byte(`---`))
		Expect(err).ToNot(HaveOccurred())

		Expect(kpis).To(HaveLen(0))
	})

	It("returns a converted kpi", func() {
		kpis, err := kpi.ReadKPIsFromYaml([]byte(`---
- name: test-kpi
  description: desc
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

		Expect(err).ToNot(HaveOccurred())

		Expect(kpis).To(Equal([]kpi.KPI{{
			Name:        "test-kpi",
			Description: "desc",
			PromQL:      "prom",
			Thresholds: []kpi.Threshold{
				{
					Level:    "warning",
					Operator: kpi.LessThan,
					Value:    0,
				},
				{
					Level:    "warning",
					Operator: kpi.LessThanOrEqualTo,
					Value:    1.2,
				},
				{
					Level:    "warning",
					Operator: kpi.EqualTo,
					Value:    0.2,
				},
				{
					Level:    "warning",
					Operator: kpi.NotEqualTo,
					Value:    123,
				},
				{
					Level:    "warning",
					Operator: kpi.GreaterThanOrEqualTo,
					Value:    642,
				},
				{
					Level:    "warning",
					Operator: kpi.GreaterThan,
					Value:    1.222225,
				},
			},
		}}))
	})

	It("returns an error if the yaml is unparsable", func() {
		_, err := kpi.ReadKPIsFromYaml([]byte(`--`))
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if a threshold has no value", func() {
		_, err := kpi.ReadKPIsFromYaml([]byte(`---
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - level: warning
  `))
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if a threshold has a bad float value", func() {
		_, err := kpi.ReadKPIsFromYaml([]byte(`---
- name: test-kpi
  description: desc
  promql: prom
  thresholds:
  - gte: abs
    level: warning
  `))
		Expect(err).To(HaveOccurred())
	})
})
