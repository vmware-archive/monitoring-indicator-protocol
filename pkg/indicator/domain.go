package indicator

import (
	"github.com/cppforlife/go-patch/patch"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

type Patch struct {
	APIVersion string
	Match      Match
	Operations []patch.OpDefinition
}

type Match struct {
	Name     *string
	Version  *string
	Metadata map[string]string
}

func GetComparatorAbbrev(op v1alpha1.ThresholdOperator) string {
	switch op {
	case v1alpha1.LessThan:
		return "lt"
	case v1alpha1.LessThanOrEqualTo:
		return "lte"
	case v1alpha1.EqualTo:
		return "eq"
	case v1alpha1.NotEqualTo:
		return "neq"
	case v1alpha1.GreaterThanOrEqualTo:
		return "gte"
	case v1alpha1.GreaterThan:
		return "gt"
	default:
		return ""
	}
}

func GetComparator(op v1alpha1.ThresholdOperator) string {
	switch op {
	case v1alpha1.LessThan:
		return "<"
	case v1alpha1.LessThanOrEqualTo:
		return "<="
	case v1alpha1.EqualTo:
		return "=="
	case v1alpha1.NotEqualTo:
		return "!="
	case v1alpha1.GreaterThanOrEqualTo:
		return ">="
	case v1alpha1.GreaterThan:
		return ">"
	}

	return ""
}

func GetComparatorFromString(operator string) v1alpha1.ThresholdOperator {
	switch operator {
	case "lt":
		return v1alpha1.LessThan
	case "lte":
		return v1alpha1.LessThanOrEqualTo
	case "eq":
		return v1alpha1.EqualTo
	case "neq":
		return v1alpha1.NotEqualTo
	case "gte":
		return v1alpha1.GreaterThanOrEqualTo
	case "gt":
		return v1alpha1.GreaterThan
	default:
		return -1
	}
}
