package tools

import (
	"os"
	"os/exec"
	"path/filepath"
)

// * Run from the root of the repo we want to analyse
// * NODE_PATH="<the installed eslint path>/node_modules"
// * The local installed ESLint should have the @microsoft/eslint-formatter-sarif installed
func RunEslint(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string, pathsToCheck []string, autoFix bool, outputFile string) {
	eslintInstallationNodeModules := filepath.Join(eslintInstallationDirectory, "node_modules")
	eslintJsPath := filepath.Join(eslintInstallationNodeModules, ".bin", "eslint")

	cmd := exec.Command(nodeBinary, eslintJsPath)
	if autoFix {
		cmd.Args = append(cmd.Args, "--fix")
	}
	if outputFile != "" {
		//When writing to file, we write is SARIF
		cmd.Args = append(cmd.Args, "-f", "@microsoft/eslint-formatter-sarif", "-o", outputFile)
	}
	if len(pathsToCheck) > 0 {
		cmd.Args = append(cmd.Args, pathsToCheck...)
	} else {
		cmd.Args = append(cmd.Args, ".")
	}

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	nodePathEnv := "NODE_PATH=" + eslintInstallationNodeModules
	cmd.Env = append(cmd.Env, nodePathEnv)

	//DEBUG
	//fmt.Println(cmd.Env)
	//fmt.Println(cmd)

	// TODO eslint returns 1 when it finds errors, so we're not propagating it
	cmd.Run()
}
