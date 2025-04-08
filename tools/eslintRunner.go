package tools

import (
	"codacy/cli-v2/config"
	"os"
	"os/exec"
	"path/filepath"
)

// * Run from the root of the repo we want to analyse
// * NODE_PATH="<the installed eslint path>/node_modules"
// * The local installed ESLint should have the @microsoft/eslint-formatter-sarif installed
func RunEslint(repositoryToAnalyseDirectory string, eslintInstallationDirectory string, nodeBinary string, pathsToCheck []string, autoFix bool, outputFile string, outputFormat string) {
	eslintInstallationNodeModules := filepath.Join(eslintInstallationDirectory, "node_modules")
	eslintJsPath := filepath.Join(eslintInstallationNodeModules, ".bin", "eslint")

	cmd := exec.Command(nodeBinary, eslintJsPath)

	// For Eslint compatibility with version 8.
	// https://eslint.org/docs/v8.x/use/configure/configuration-files-new
	cmd.Env = append(cmd.Env, "ESLINT_USE_FLAT_CONFIG=true")

	// Add config file from tools-configs directory
	configFile := filepath.Join(config.Config.ToolsConfigDirectory(), "eslint.config.mjs")
	cmd.Args = append(cmd.Args, "-c", configFile)

	if autoFix {
		cmd.Args = append(cmd.Args, "--fix")
	}
	if outputFormat == "sarif" {
		//When outputting in SARIF format
		cmd.Args = append(cmd.Args, "-f", "@microsoft/eslint-formatter-sarif")
	}

	if outputFile != "" {
		//When writing to file, use the output file option
		cmd.Args = append(cmd.Args, "-o", outputFile)
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

	// DEBUG
	// fmt.Println(cmd.Env)
	// fmt.Println(cmd)

	// TODO eslint returns 1 when it finds errors, so we're not propagating it
	cmd.Run()
}
