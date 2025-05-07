package tools

import (
	"codacy/cli-v2/config"
	"fmt"
	"os"
	"os/exec"
)

// RunTrivy executes Trivy vulnerability scanner with the specified options
func RunTrivy(repositoryToAnalyseDirectory string, trivyBinary string, pathsToCheck []string, outputFile string, outputFormat string) error {
	cmd := exec.Command(trivyBinary, "fs")

	// Add config file from tools-configs directory if it exists
	if configFile, exists := ConfigFileExists(config.Config, "trivy.yaml"); exists {
		cmd.Args = append(cmd.Args, "--config", configFile)
	}

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

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run Trivy: %w", err)
	}
	return nil
}
