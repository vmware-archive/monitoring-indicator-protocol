package go_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

func Build(packagePath string, args ...string) (compiledPath string, err error) {
	tmpDir, err := ioutil.TempDir("", "gexec_artifacts")
	if err != nil {
		return "", err
	}

	cmdArgs := append(append([]string{"build", "-mod=vendor", "-o", tmpDir+"/build"}, args...), packagePath)
	build := exec.Command("go", cmdArgs...)

	output, err := build.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to build %s:\n\nError:\n%s\n\nOutput:\n%s", packagePath, err, string(output))
	}

	return tmpDir + "/build", nil
}
