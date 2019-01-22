package configuration

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	glob2 "github.com/gobwas/glob"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"
)

type GitGetter func(Source) ([]indicator.Patch, []indicator.Document, error)

var getGitPatchesAndDocuments GitGetter = realGitGet

func SetGitGetter(getter GitGetter) {
	getGitPatchesAndDocuments = getter
}

func Read(configFile string) ([]indicator.Patch, []indicator.Document, error) {
	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading configuration file: %s\n", err)
	}

	var f File
	err = yaml.Unmarshal(fileBytes, &f)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing configuration file: %s\n", err)
	}

	if err := Validate(f); err != nil {
		return nil, nil, fmt.Errorf("configuration is not valid: %s\n", err)
	}

	var patches []indicator.Patch
	var documents []indicator.Document
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
			gitPatches, gitDocuments, err := getGitPatchesAndDocuments(source)
			if err != nil {
				log.Printf("failed to read patches in %s from config file %s: %s\n", source.Repository, configFile, err)
				continue
			}
			patches = append(patches, gitPatches...)
			documents = append(documents, gitDocuments...)
		default:
			log.Printf("invalid type [%s] in file: %s\n", source.Type, configFile)
			continue
		}
	}

	return patches, documents, nil
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
	Glob       string `yaml:"glob"`
}

func realGitGet(s Source) ([]indicator.Patch, []indicator.Document, error) {
	storage := memory.NewStorage()

	var auth transport.AuthMethod = nil
	if s.Token != "" {
		auth = &http.BasicAuth{
			Username: "github",
			Password: s.Token,
		}
	}

	repo := s.Repository
	r, err := git.Clone(storage, nil, &git.CloneOptions{
		Auth: auth,
		URL:  repo,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to clone repo: %s\n", err)
	}

	ref, err := r.Head()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch repo head: %s\n", err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch commit: %s\n", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch commit tree: %s\n", err)
	}

	return retrievePatchesAndDocuments(tree.Files(), repo, s.Glob)
}

func retrievePatchesAndDocuments(files *object.FileIter, repo string, glob string) ([]indicator.Patch, []indicator.Document, error) {
	var patchesBytes []unparsedPatch
	var documentsBytes [][]byte

	if glob == "" {
		glob = "*.y*ml"
	}
	g := glob2.MustCompile(glob)

	err := files.ForEach(func(f *object.File) error {
		if g.Match(f.Name) {
			contents, err := f.Contents()
			if err != nil {
				log.Println(err)
				return nil
			}

			apiVersion := getAPIVersion([]byte(contents))
			switch apiVersion {
			case "v1/patch":
				patchesBytes = append(patchesBytes, unparsedPatch{[]byte(contents), f.Name})
			case "v0":
				documentsBytes = append(documentsBytes, []byte(contents))
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to traverse git tree: %s\n", err)
	}
	patches := readPatches(patchesBytes, repo)
	documents := processDocuments(documentsBytes, patches)
	return patches, documents, err
}

type unparsedPatch struct {
	YAMLBytes []byte
	Filename  string
}

func readPatches(unparsedPatches []unparsedPatch, repo string) []indicator.Patch {
	var patches []indicator.Patch
	for _, p := range unparsedPatches {
		p, err := indicator.ReadPatchBytes(fmt.Sprintf("%v %v", repo, p.Filename), p.YAMLBytes)
		if err != nil {
			log.Println(err)
		}
		patches = append(patches, p)
	}
	return patches
}

func processDocuments(documentsBytes [][]byte, patches []indicator.Patch) []indicator.Document {
	var documents []indicator.Document
	for _, documentBytes := range documentsBytes {
		doc, errs := indicator.ProcessDocument(patches, documentBytes)
		if len(errs) > 0 {
			for _, e := range errs {
				log.Printf("- %s \n", e.Error())
			}

			log.Printf("validation for indicator file failed - [%+v]\n", errs)
		}

		documents = append(documents, doc)
	}
	return documents
}

func getAPIVersion(fileBytes []byte) string {
	var f struct {
		APIVersion string `yaml:"apiVersion"`
	}

	err := yaml.Unmarshal(fileBytes, &f)
	if err != nil {
		return err.Error()
	}

	return f.APIVersion
}
