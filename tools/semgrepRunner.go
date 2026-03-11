package tools

import (
	"codacy/cli-v2/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunOpengrep executes Opengrep analysis on the specified directory
func RunOpengrep(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	cmdArgs := []string{"scan"}

	cmdArgs = append(cmdArgs, "--max-memory", "2560")
	cmdArgs = append(cmdArgs, "--timeout", "5")
	cmdArgs = append(cmdArgs, "--timeout-threshold", "3")

	cmdArgs = append(cmdArgs, "--disable-version-check")

	// Add output format if specified
	if outputFormat == "sarif" {
		cmdArgs = append(cmdArgs, "--sarif")
	}

	// Define possible Opengrep config file names (uses semgrep-compatible config format)
	opengrepConfigFiles := []string{"semgrep.yml", "semgrep.yaml", "semgrep/semgrep.yml"}

	// Check if a config file exists in the expected location and use it if present
	if configFile, exists := ConfigFileExists(config.Config, opengrepConfigFiles...); exists {
		cmdArgs = append(cmdArgs, "--config", configFile)
	} else {
		// add --config auto only if no config file exists
		cmdArgs = append(cmdArgs, "--config", "auto")
	}

	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		cmdArgs = append(cmdArgs, files...)
	} else {
		cmdArgs = append(cmdArgs, ".")
	}

	// Create Opengrep command
	cmd := exec.Command(binary, cmdArgs...)
	cmd.Dir = workDirectory

	if outputFile != "" {
		// If output file is specified, create it and redirect output
		var outputWriter *os.File
		var err error
		outputWriter, err = os.Create(filepath.Clean(outputFile))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.Close()
		cmd.Stdout = outputWriter
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	// Run Opengrep
	if err := cmd.Run(); err != nil {
		// Opengrep returns non-zero exit code when it finds issues, which is expected
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("failed to run opengrep: %w", err)
		}
	}

	return nil
}
