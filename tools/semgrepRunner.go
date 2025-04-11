package tools

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/plugins"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunSemgrep executes Semgrep analysis on the specified directory
func RunSemgrep(workDirectory string, toolInfo *plugins.ToolInfo, files []string, outputFile string, outputFormat string) error {
	// Construct base command with -m semgrep to run semgrep module
	cmdArgs := []string{"scan"}

	// Add output format if specified
	if outputFormat == "sarif" {
		cmdArgs = append(cmdArgs, "--sarif")
	}

	// Check if a config file exists in the expected location and use it if present
	if configFile, exists := ConfigFileExists(config.Config, ".semgrep.yml"); exists {
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

	cmdArgs = append(cmdArgs, "--disable-version-check")

	// Get Semgrep binary from the specified installation path
	semgrepPath := filepath.Join(toolInfo.InstallDir, "venv", "bin", "semgrep")

	// Create Semgrep command
	cmd := exec.Command(semgrepPath, cmdArgs...)
	cmd.Dir = workDirectory

	// If output file is specified, create it and redirect output
	var outputWriter *os.File
	var err error
	if outputFile != "" {
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

	// Run Semgrep
	if err := cmd.Run(); err != nil {
		// Semgrep returns non-zero exit code when it finds issues, which is expected
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("failed to run semgrep: %w", err)
		}
	}

	return nil
}
