package asset_test

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	. "github.com/onsi/gomega"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/asset"
)

func TestIndicatorDocumentSchema(t *testing.T) {
	g := NewGomegaWithT(t)
	schemaBytes, err := asset.Asset("schemas.yml")
	g.Expect(err).To(BeNil())

	var schema struct {
		IndicatorDocumentSchema spec.Schema `json:"IndicatorDocument"`
	}
	var rootSchema interface{}
	err = yaml.Unmarshal(schemaBytes, &rootSchema)
	err = yaml.Unmarshal(schemaBytes, &schema)
	validator := validate.NewSchemaValidator(&schema.IndicatorDocumentSchema, rootSchema, "IndicatorDocument", strfmt.Default)

	g.Expect(err).To(BeNil())

	t.Run("Accepts the example indicator document", func(t *testing.T) {
		var exampleDocument map[string]interface{}

		exampleDocBytes, err := ioutil.ReadFile("../../example_indicators.yml")
		g.Expect(err).To(BeNil())
		err = yaml.Unmarshal(exampleDocBytes, &exampleDocument)
		g.Expect(err).To(BeNil())

		g.Expect(validator.Validate(exampleDocument).IsValid()).To(BeTrue())
	})

	t.Run("Does not accept invalid indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)
		var indicator interface{}
		err := yaml.Unmarshal([]byte(`yaml: invalidindicator`), &indicator)
		g.Expect(err).To(BeNil())
		g.Expect(validator.Validate(indicator).IsValid()).To(BeFalse())
	})
}
