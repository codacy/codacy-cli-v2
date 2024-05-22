package tools

import (
	"os"
	"os/exec"
	"path/filepath"
)

func RunEslintToFile(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string, outputFolder string) error {
	_, err := runEslint(repositoryToAnalyseDirectory, eslintInstallationDirectory, nodeBinary, outputFolder)
	return err
}

func RunEslintToString(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string) (string, error) {
	return runEslint(repositoryToAnalyseDirectory, eslintInstallationDirectory, nodeBinary, "")
}

// * Run from the root of the repo we want to analyse
// * NODE_PATH="<the installed eslint path>/node_modules"
// * The local installed ESLint should have the @microsoft/eslint-formatter-sarif installed
func runEslint(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string, outputFolder string) (string, error) {
	eslintInstallationNodeModules := filepath.Join(eslintInstallationDirectory, "node_modules")
	eslintJsPath := filepath.Join(eslintInstallationNodeModules, ".bin", "eslint")

	cmd := exec.Command(nodeBinary, eslintJsPath, "-f", "@microsoft/eslint-formatter-sarif")

	if outputFolder != "" {
		outputFile := filepath.Join(outputFolder, "eslint.sarif")
		cmd.Args = append(cmd.Args, "-o", outputFile)
	}

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr

	nodePathEnv := "NODE_PATH=" + eslintInstallationNodeModules
	cmd.Env = append(cmd.Env, nodePathEnv)

	// TODO eslint returns 1 when it finds errors, so we're not propagating it
	out, _ := cmd.Output()

	return string(out), nil
}
