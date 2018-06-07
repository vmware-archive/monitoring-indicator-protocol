package vgo_test

import (
	"os/exec"
	"fmt"
)

func Build(packagePath string, args ...string) (compiledPath string, err error) {
	build := exec.Command("vgo", "build", "-o", "/tmp/build", packagePath)

	output, err := build.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to build %s:\n\nError:\n%s\n\nOutput:\n%s", packagePath, err, string(output))
	}

	return "/tmp/build", nil
}
