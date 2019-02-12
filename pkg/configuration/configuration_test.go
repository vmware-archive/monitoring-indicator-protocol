package configuration_test

import (
	"bytes"
	"log"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotal/indicator-protocol/pkg/configuration"
	"github.com/pivotal/indicator-protocol/pkg/go_test"
	"gopkg.in/src-d/go-git.v4"
)

func TestReadLocalConfigurationFile(t *testing.T) {
	g := NewGomegaWithT(t)

	sourceFile, _ := configuration.ParseSourcesFile("test_fixtures/local_config.yml")

	patches, _, err := configuration.Read(sourceFile, nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(patches).To(HaveLen(2))

	patch1 := patches[0].Patches[0]
	patch2 := patches[1].Patches[0]

	g.Expect(*patch1.Match.Name).To(Equal("my-component-1"))
	g.Expect(*patch2.Match.Name).To(Equal("my-component-2"))
}

func TestReadGitConfigurationFile(t *testing.T) {
	g := NewGomegaWithT(t)

	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	fakeRepository := go_test.CreateMemoryRepo(
		"test_fixtures/patch1.yml",
		"test_fixtures/patch2.yml",
		"test_fixtures/indicators1.yml",
		"test_fixtures/indicators2.yml")

	fakeGetter := func(s configuration.Source) (*git.Repository, error) {
		return fakeRepository, nil
	}

	sources, _ := configuration.ParseSourcesFile("test_fixtures/git_config.yml")

	patches, documents, err := configuration.Read(sources, fakeGetter)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(patches).To(HaveLen(1))
	g.Expect(patches[0].Source).To(Equal("https://fakegit.nope/slowens/test-repo.git"))
	g.Expect(patches[0].Patches).To(HaveLen(2))
	g.Expect(*patches[0].Patches[0].Match.Name).To(Equal("my-component-1"))
	g.Expect(*patches[0].Patches[1].Match.Name).To(Equal("my-component-2"))

	g.Expect(documents).To(HaveLen(2))
	g.Expect(documents[0].Product.Name).To(Equal("my-component"))
	g.Expect(documents[1].Product.Name).To(Equal("someone-elses-component"))

	g.Expect(buffer.String()).To(ContainSubstring("Parsed 2 documents and 2 patches from https://fakegit.nope/slowens/test-repo.git git source"))
}

func TestGlobMatching(t *testing.T) {
	g := NewGomegaWithT(t)

	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	fakeRepository := go_test.CreateMemoryRepo(
		"test_fixtures/patch1.yml",
		"test_fixtures/patch2.yml",
		"test_fixtures/indicators1.yml",
		"test_fixtures/indicators2.yml",
		"test_fixtures/bad.yml")

	fakeGetter := func(s configuration.Source) (*git.Repository, error) {
		return fakeRepository, nil
	}

	_, _, err := configuration.Read([]configuration.Source{{
		Type:       "git",
		Repository: "fake/fake/fake",
		Glob:       "patch*.yml",
	}}, fakeGetter)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(buffer.String()).To(ContainSubstring("Parsed 0 documents and 2 patches from fake/fake/fake git source"))
}

func TestValidateConfigFile(t *testing.T) {
	t.Run("does not return error if token is not provided with SSH git repo", func(t *testing.T) {
		g := NewGomegaWithT(t)

		err := configuration.Validate(configuration.SourcesFile{
			Sources: []configuration.Source{{
				Type:       "git",
				Repository: "git@fakegit.nope:slowens/test-repo.git",
			}},
		})
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("returns error if token is provided with SSH git repo", func(t *testing.T) {
		g := NewGomegaWithT(t)

		err := configuration.Validate(configuration.SourcesFile{
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

		err := configuration.Validate(configuration.SourcesFile{
			Sources: []configuration.Source{{
				Type: "git",
			}},
		})
		g.Expect(err).To(MatchError(ContainSubstring("repository is required for git sources")))
	})

	t.Run("returns error if path isn't provided in local source", func(t *testing.T) {
		g := NewGomegaWithT(t)

		err := configuration.Validate(configuration.SourcesFile{
			Sources: []configuration.Source{{
				Type: "local",
			}},
		})
		g.Expect(err).To(MatchError(ContainSubstring("path is required for local sources")))
	})
}

func TestFailToParseConfigurationFile(t *testing.T) {
	t.Run("returns an error if config file cannot be read", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := configuration.ParseSourcesFile(`files are overrated`)
		g.Expect(err).To(MatchError(ContainSubstring("error reading configuration file:")))
	})

	t.Run("returns an error if config cannot be parsed", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, err := configuration.ParseSourcesFile("test_fixtures/bad.yml")
		g.Expect(err).To(MatchError(ContainSubstring("error parsing configuration file:")))
	})

	t.Run("returns a partial list if some patches cannot be read", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		sourceFile, _ := configuration.ParseSourcesFile("test_fixtures/partial_bad.yml")

		patches, _, err := configuration.Read(sourceFile, nil)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(patches).To(HaveLen(1))
		g.Expect(*patches[0].Patches[0].Match.Name).To(Equal("my-component-1"))

		g.Expect(buffer.String()).To(ContainSubstring("failed to read local patch badpath/nothing_here.yml"))
	})
}
