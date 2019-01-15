package configuration_test

import (
	. "github.com/onsi/gomega"
	"testing"

	"code.cloudfoundry.org/indicators/pkg/configuration"
	"code.cloudfoundry.org/indicators/pkg/indicator"
)

func TestReadLocalConfigurationFile(t *testing.T) {
	g := NewGomegaWithT(t)

	patches, _, err := configuration.Read("test_fixtures/local_config.yml")
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(patches).To(HaveLen(2))

	patch1 := patches[0]
	patch2 := patches[1]

	g.Expect(*patch1.Match.Name).To(Equal("my-component-1"))
	g.Expect(*patch2.Match.Name).To(Equal("my-component-2"))
}

func TestReadGitConfigurationFile(t *testing.T) {
	g := NewGomegaWithT(t)
	testPatches := []indicator.Patch{{
		Origin:     "test-file",
		APIVersion: "whocares",
	}}

	testDocuments := []indicator.Document{{
		APIVersion: "whocares-doc",
	}}

	fakeGetter := func(s configuration.Source) ([]indicator.Patch, []indicator.Document, error) {
		g.Expect(s.Repository).To(Equal("https://fakegit.nope/slowens/test-repo.git"))
		g.Expect(s.Token).To(Equal("test_private_key"))
		return testPatches, testDocuments, nil
	}

	configuration.SetGitGetter(fakeGetter)

	patches, documents, err := configuration.Read("test_fixtures/git_config.yml")
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(patches).To(ConsistOf(testPatches))
	g.Expect(documents).To(ConsistOf(testDocuments))
}

func TestValidateConfigFile(t *testing.T) {
	t.Run("does not return error if token is not provided with SSH git repo", func(t *testing.T) {
		g := NewGomegaWithT(t)

		err := configuration.Validate(configuration.File{
			Sources: []configuration.Source{{
				Type:       "git",
				Repository: "git@fakegit.nope:slowens/test-repo.git",
			}},
		})
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("returns error if token is provided with SSH git repo", func(t *testing.T) {
		g := NewGomegaWithT(t)

		err := configuration.Validate(configuration.File{
			Sources: []configuration.Source{{
				Type:       "git",
				Repository: "git@fakegit.nope:slowens/test-repo.git",
				Token:      "asdfasdf",
			}},
		})
		g.Expect(err).To(MatchError(ContainSubstring("personal access token can only be used over HTTPS")))
	})

	t.Run("returns error if repo isn't provided in git source", func(t *testing.T) {
		g := NewGomegaWithT(t)

		err := configuration.Validate(configuration.File{
			Sources: []configuration.Source{{
				Type: "git",
			}},
		})
		g.Expect(err).To(MatchError(ContainSubstring("repository is required for git sources")))
	})

	t.Run("returns error if path isn't provided in local source", func(t *testing.T) {
		g := NewGomegaWithT(t)

		err := configuration.Validate(configuration.File{
			Sources: []configuration.Source{{
				Type: "local",
			}},
		})
		g.Expect(err).To(MatchError(ContainSubstring("path is required for local sources")))
	})
}

func TestFailToReadConfigurationFile(t *testing.T) {
	t.Run("returns an error if config file cannot be read", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, _, err := configuration.Read(`files are overrated`)
		g.Expect(err).To(MatchError(ContainSubstring("error reading configuration file:")))
	})

	t.Run("returns an error if config cannot be parsed", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, _, err := configuration.Read("test_fixtures/bad.yml")
		g.Expect(err).To(MatchError(ContainSubstring("error parsing configuration file:")))
	})

	t.Run("returns a partial list if some patches cannot be read", func(t *testing.T) {
		g := NewGomegaWithT(t)

		patches, _, err := configuration.Read("test_fixtures/partial_bad.yml")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(patches).To(HaveLen(1))
		g.Expect(*patches[0].Match.Name).To(Equal("my-component-1"))
	})
}
