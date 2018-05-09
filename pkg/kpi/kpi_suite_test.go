package kpi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKpi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kpi Suite")
}
