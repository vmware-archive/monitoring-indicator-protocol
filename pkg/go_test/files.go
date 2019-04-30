package go_test

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/src-d/go-billy.v4"
)

func WaitForFiles(directory string, count int) error {
	for range [100]int{} {
		files, err := ioutil.ReadDir(directory)
		if err != nil {
			return err
		}

		if len(files) >= count {
			file, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, files[0].Name()))
			if err != nil {
				return err
			}
			if len(file) > 0 {
				return nil
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("files not found in %s", directory)
}

func GetFileNames(fs billy.Filesystem, directory string) ([]string, error) {
	files, err := fs.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	fileNames := make([]string, 0)
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
}
