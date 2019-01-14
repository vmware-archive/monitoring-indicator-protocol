package configuration

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"
)

type GitPatchGetter func(Source) ([]indicator.Patch, error)

var getGitPatches GitPatchGetter = realGetGitPatches

func SetGitPatchGetter(getter GitPatchGetter) {
	getGitPatches = getter
}

func Read(configFile string) ([]indicator.Patch, error) {
	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s\n", err)
	}

	var f File
	err = yaml.Unmarshal(fileBytes, &f)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %s\n", err)
	}

	if err := Validate(f); err != nil {
		return nil, fmt.Errorf("configuration is not valid: %s\n", err)
	}

	var patches []indicator.Patch
	for _, source := range f.Sources {

		switch source.Type {
		case "local":
			patch, err := indicator.ReadPatchFile(source.Path)
			if err != nil {
				log.Printf("failed to read patch %s from config file %s: %s\n", source.Path, configFile, err)
				continue
			}

			patches = append(patches, patch)
		case "git":
			gitPatches, err := getGitPatches(source)
			if err != nil {
				log.Printf("failed to read patches in %s from config file %s: %s\n", source.Repository, configFile, err)
				continue
			}
			patches = append(patches, gitPatches...)
		default:
			log.Printf("invalid type [%s] in file: %s\n", source.Type, configFile)
			continue
		}
	}

	return patches, nil
}

func realGetGitPatches(s Source) ([]indicator.Patch, error) {
	storage := memory.NewStorage()

	var auth transport.AuthMethod = nil
	if s.Token != "" {
		auth = &http.BasicAuth{
			Username: "github",
			Password: s.Token,
		}
	}

	r, err := git.Clone(storage, nil, &git.CloneOptions{
		Auth: auth,
		URL:  s.Repository,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repo: %s\n", err)
	}

	ref, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo head: %s\n", err)
	}

	commit, err := r.CommitObject(ref.Hash())
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit: %s\n", err)
	}

	var patches []indicator.Patch
	err = tree.Files().ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, ".yml") {
			contents, err := f.Contents()
			if err != nil {
				log.Println(err)
				return nil
			}

			p, err := indicator.ReadPatchBytes(fmt.Sprintf("%s/%s", s.Repository, f.Name), []byte(contents))
			if err != nil {
				log.Println(err)
				return nil
			}

			if p.APIVersion == "v1/patch" {
				patches = append(patches, p)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to traverse git tree: %s\n", err)
	}

	return patches, nil
}

func Validate(f File) error {
	for _, s := range f.Sources {
		switch s.Type {
		case "local":
			if s.Path == "" {
				return fmt.Errorf("path is required for local sources")
			}
		case "git":
			if s.Repository == "" {
				return fmt.Errorf("repository is required for git sources")
			}

			if s.Token != "" && !strings.Contains(s.Repository, "https://") {
				return fmt.Errorf("personal access token can only be used over HTTPS")
			}
		}
	}

	return nil
}

type File struct {
	Sources []Source `yaml:"sources"`
}

type Source struct {
	Type       string `yaml:"type"`
	Path       string `yaml:"path"`
	Repository string `yaml:"repository"`
	Token      string `yaml:"token"`
}
