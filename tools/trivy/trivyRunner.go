package trivy

import (
	"codacy/cli-v2/config"
	"os"
	"os/exec"
	"path/filepath"
)

// RunTrivy executes Trivy vulnerability scanner with the specified options
func RunTrivy(repositoryToAnalyseDirectory string, trivyBinary string, pathsToCheck []string, outputFile string, outputFormat string) error {
	cmd := exec.Command(trivyBinary, "fs")

	// Add config file from tools-configs directory
	configFile := filepath.Join(config.Config.ToolsConfigDirectory(), "trivy.yaml")
	cmd.Args = append(cmd.Args, "--config", configFile)

	// Add format options
	if outputFile != "" {
		cmd.Args = append(cmd.Args, "--output", outputFile)
	}

	if outputFormat == "sarif" {
		cmd.Args = append(cmd.Args, "--format", "sarif")
	}

	// Add specific targets or use current directory
	if len(pathsToCheck) > 0 {
		cmd.Args = append(cmd.Args, pathsToCheck...)
	} else {
		cmd.Args = append(cmd.Args, ".")
	}

	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
