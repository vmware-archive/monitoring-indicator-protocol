package configuration_test

import (
	"bytes"
	"fmt"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/util"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"io/ioutil"
	"log"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/pivotal/indicator-protocol/pkg/configuration"
)

func TestReadLocalConfigurationFile(t *testing.T) {
	g := NewGomegaWithT(t)

	patches, _, err := configuration.Read("test_fixtures/local_config.yml", nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(patches).To(HaveLen(2))

	patch1 := patches[0].Patches[0]
	patch2 := patches[1].Patches[0]

	g.Expect(*patch1.Match.Name).To(Equal("my-component-1"))
	g.Expect(*patch2.Match.Name).To(Equal("my-component-2"))
}

func TestReadGitConfigurationFile(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRepository := createTestingRepo(
		"test_fixtures/patch1.yml",
		"test_fixtures/patch2.yml",
		"test_fixtures/indicators1.yml",
		"test_fixtures/indicators2.yml")

	fakeGetter := func(s configuration.Source) (*git.Repository, error) {
		return fakeRepository, nil
	}

	patches, documents, err := configuration.Read("test_fixtures/git_config.yml", fakeGetter)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(patches).To(HaveLen(1))
	g.Expect(patches[0].Source).To(Equal("https://fakegit.nope/slowens/test-repo.git"))
	g.Expect(patches[0].Patches).To(HaveLen(2))
	g.Expect(*patches[0].Patches[0].Match.Name).To(Equal("my-component-1"))
	g.Expect(*patches[0].Patches[1].Match.Name).To(Equal("my-component-2"))

	g.Expect(documents).To(HaveLen(2))
	g.Expect(documents[0].Product.Name).To(Equal("my-component"))
	g.Expect(documents[1].Product.Name).To(Equal("someone-elses-component"))

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

		_, _, err := configuration.Read(`files are overrated`, nil)
		g.Expect(err).To(MatchError(ContainSubstring("error reading configuration file:")))
	})

	t.Run("returns an error if config cannot be parsed", func(t *testing.T) {
		g := NewGomegaWithT(t)

		_, _, err := configuration.Read("test_fixtures/bad.yml", nil)
		g.Expect(err).To(MatchError(ContainSubstring("error parsing configuration file:")))
	})

	t.Run("returns a partial list if some patches cannot be read", func(t *testing.T) {
		buffer := bytes.NewBuffer(nil)
		log.SetOutput(buffer)

		g := NewGomegaWithT(t)

		patches, _, err := configuration.Read("test_fixtures/partial_bad.yml", nil)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(patches).To(HaveLen(1))
		g.Expect(*patches[0].Patches[0].Match.Name).To(Equal("my-component-1"))

		g.Expect(buffer.String()).To(ContainSubstring("failed to read patch badpath/nothing_here.yml from config file test_fixtures/partial_bad.yml"))
	})
}

func createTestingRepo(files ...string) *git.Repository {
	storage := memory.NewStorage()
	fs := memfs.New()

	repo, err := git.Init(storage, fs)
	if err != nil {
		panic(fmt.Sprintf("could not create repo: %s", err))
	}

	w, _ := repo.Worktree()

	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			panic(fmt.Errorf("could not read file '%s': %s", f, err))
		}

		_ = util.WriteFile(fs, f, data, 0644)

		_, err = w.Add(f)
		if err != nil {
			panic(fmt.Errorf("could not add file '%s' to test repository: %s", f, err))
		}
	}

	_, err = w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	if err != nil {
		panic(fmt.Errorf("could not create commit: %s", err))
	}

	return repo
}
