package configuration

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	glob2 "github.com/gobwas/glob"
	"github.com/pivotal/indicator-protocol/pkg/indicator"
	"github.com/pivotal/indicator-protocol/pkg/registry"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/yaml.v2"
)

type RepositoryGetter func(Source) (*git.Repository, error)

func Read(configFile string, repositoryGetter RepositoryGetter) ([]registry.PatchList, []indicator.Document, error) {
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

	var patches []registry.PatchList
	var documents []indicator.Document
	for _, source := range f.Sources {

		switch source.Type {
		case "local":
			patch, err := indicator.ReadPatchFile(source.Path)
			if err != nil {
				log.Printf("failed to read patch %s from config file %s: %s\n", source.Path, configFile, err)
				continue
			}

			patches = append(patches, registry.PatchList{
				Source:  source.Path,
				Patches: []indicator.Patch{patch},
			})
		case "git":
			repository, err := repositoryGetter(source)
			//TODO Improve logging: Hard to tell if the app is working from the logs right now
			if err != nil {
				log.Printf("failed to initialize repository in %s from config file %s: %s\n", source.Repository, configFile, err)
				continue
			}
			gitPatches, gitDocuments, err := realGitGet(source, repository)
			if err != nil {
				log.Printf("failed to read patches in %s from config file %s: %s\n", source.Repository, configFile, err)
				continue
			}
			patches = append(patches, registry.PatchList{
				Source:  source.Repository,
				Patches: gitPatches,
			})
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

func realGitGet(s Source, r *git.Repository) ([]indicator.Patch, []indicator.Document, error) {
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

	return retrievePatchesAndDocuments(tree.Files(), s.Glob)
}

func retrievePatchesAndDocuments(files *object.FileIter, glob string) ([]indicator.Patch, []indicator.Document, error) {
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
			case "v0/patch":
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
	patches := readPatches(patchesBytes)
	documents := processDocuments(documentsBytes, patches)
	return patches, documents, err
}

type unparsedPatch struct {
	YAMLBytes []byte
	Filename  string
}

func readPatches(unparsedPatches []unparsedPatch) []indicator.Patch {
	var patches []indicator.Patch
	for _, p := range unparsedPatches {
		p, err := indicator.ReadPatchBytes(p.YAMLBytes)
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
			continue
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
