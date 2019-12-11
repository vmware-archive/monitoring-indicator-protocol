package go_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-billy.v4/util"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func CreateMemoryRepo(files ...string) *git.Repository {
	memStorage := memory.NewStorage()
	fs := memfs.New()

	return createRepo(memStorage, fs, files)
}

func CreateTempRepo(files ...string) string {
	rootPath, err := ioutil.TempDir("", "test_repo")
	if err != nil {
		panic(fmt.Errorf("could not create temp dir: %s", err))
	}
	fs := osfs.New(rootPath)
	dot, _ := fs.Chroot(rootPath)
	tempStorage := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	createRepo(tempStorage, fs, files)

	return rootPath
}

func createRepo(storage storage.Storer, fs billy.Filesystem, files []string) *git.Repository {
	repo, err := git.Init(storage, fs)
	if err != nil {
		panic(fmt.Sprintf("could not create repo: %s", err))
	}

	w, _ := repo.Worktree()

	for _, f := range files {
		fileBytes, err := ioutil.ReadFile(f)
		if err != nil {
			panic(fmt.Errorf("could not read file '%s': %s", f, err))
		}

		err = fs.MkdirAll(filepath.Dir(f), 0666)
		if err != nil {
			panic(fmt.Errorf("could not make dir '%s': %s", filepath.Dir(f), err))
		}

		writeFile := f

		_ = util.WriteFile(fs, writeFile, fileBytes, 0644)

		_, err = w.Add(writeFile)
		if err != nil {
			panic(fmt.Errorf("could not add file '%s' to test repository: %s", writeFile, err))
		}
	}

	_, err = w.Commit("good commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@pivotal.io",
			When:  time.Now(),
		},
	})
	if err != nil {
		panic(fmt.Errorf("could not create commit: %s", err))
	}

	return repo
}
