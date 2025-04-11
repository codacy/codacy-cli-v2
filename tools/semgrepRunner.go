package tools

import (
	"codacy/cli-v2/plugins"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunSemgrep executes Semgrep analysis on the specified directory
func RunSemgrep(workDirectory string, toolInfo *plugins.ToolInfo, files []string, outputFile string, outputFormat string) error {
	// Get Python binary from venv
	pythonPath := filepath.Join(toolInfo.InstallDir, "venv", "bin", "python3")

	// Construct base command with -m semgrep to run semgrep module
	cmdArgs := []string{"-m", "semgrep", "scan"}

	// Add output format if specified
	if outputFormat == "sarif" {
		cmdArgs = append(cmdArgs, "--sarif")
	}

	// add --config auto
	cmdArgs = append(cmdArgs, "--config", "auto")

	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		cmdArgs = append(cmdArgs, files...)
	} else {
		cmdArgs = append(cmdArgs, ".")
	}

	// Create Semgrep command
	cmd := exec.Command(pythonPath, cmdArgs...)
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
