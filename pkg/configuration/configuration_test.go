package configuration_test

import (
	. "github.com/onsi/gomega"
	"testing"

	"code.cloudfoundry.org/indicators/pkg/configuration"
)

func TestReadConfigurationFile(t *testing.T) {
	g := NewGomegaWithT(t)

	patches, err := configuration.Read("test_fixtures/config.yml")
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(patches).To(HaveLen(2))

	patch1 := patches[0]
	patch2 := patches[1]

	g.Expect(*patch1.Match.Name).To(Equal("my-component-1"))
	g.Expect(*patch2.Match.Name).To(Equal("my-component-2"))
}


func TestFailToReadConfigurationFile(t *testing.T) {
	t.Run("returns an error if config file cannot be read", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := configuration.Read(`files are overrated`)
		g.Expect(err).To(MatchError(ContainSubstring("error reading configuration file:")))
	})

	t.Run("returns an error if config cannot be parsed", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := configuration.Read("test_fixtures/bad.yml")
		g.Expect(err).To(MatchError(ContainSubstring("error parsing configuration file:")))
	})

	t.Run("returns a partial list if some patches cannot be read", func(t *testing.T) {
	    g:= NewGomegaWithT(t)

		patches, err := configuration.Read("test_fixtures/partial_bad.yml")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(patches).To(HaveLen(1))
		g.Expect(*patches[0].Match.Name).To(Equal("my-component-1"))
	})
}
