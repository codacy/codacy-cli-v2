package tools

import (
	"os"
	"os/exec"
	"path/filepath"
)

func RunEslintToFile(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string, outputFile string) error {
	_, err := runEslint(repositoryToAnalyseDirectory, eslintInstallationDirectory, nodeBinary, outputFile)
	return err
}

func RunEslintToString(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string) (string, error) {
	return runEslint(repositoryToAnalyseDirectory, eslintInstallationDirectory, nodeBinary, "")
}

// * Run from the root of the repo we want to analyse
// * NODE_PATH="<the installed eslint path>/node_modules"
// * The local installed ESLint should have the @microsoft/eslint-formatter-sarif installed
func runEslint(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string, outputFile string) (string, error) {
	eslintInstallationNodeModules := filepath.Join(eslintInstallationDirectory, "node_modules")
	eslintJsPath := filepath.Join(eslintInstallationNodeModules, ".bin", "eslint")

	cmd := exec.Command(nodeBinary, eslintJsPath)
	if outputFile != "" {
		//When writing to file, we write is SARIF
		cmd.Args = append(cmd.Args, "-f", "@microsoft/eslint-formatter-sarif", "-o", outputFile)
	}

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr

	nodePathEnv := "NODE_PATH=" + eslintInstallationNodeModules
	cmd.Env = append(cmd.Env, nodePathEnv)

	// TODO eslint returns 1 when it finds errors, so we're not propagating it
	out, _ := cmd.Output()

	//DEBUG:
	//fmt.Println(cmd.Env)
	//fmt.Println(cmd)

	return string(out), nil
}
