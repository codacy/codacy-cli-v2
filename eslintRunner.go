package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// * Run from the root of the repo we want to analyse
// * NODE_PATH="<the installed eslint path>/node_modules"
// * The local installed ESLint should have the @microsoft/eslint-formatter-sarif installed
func runEslint(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string) (string, error) {
	eslintInstallationNodeModules := filepath.Join(eslintInstallationDirectory, "node_modules")
	eslintJsPath := filepath.Join(eslintInstallationNodeModules, ".bin/eslint")

	cmd := exec.Command(nodeBinary, eslintJsPath, "-f", "@microsoft/eslint-formatter-sarif", ".")
	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr

	fmt.Println(repositoryToAnalyseDirectory)

	nodePathEnv := "NODE_PATH=" + eslintInstallationNodeModules
	cmd.Env = append(cmd.Env, nodePathEnv)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
