package tools

import (
	"codacy/cli-v2/utils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// RunPyrefly executes Pyrefly type checking on the specified directory or files
func RunPyrefly(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	args := []string{"check"}

	// Always use JSON output for SARIF conversion
	var tempFile string
	if outputFormat == "sarif" {
		tmp, err := ioutil.TempFile("", "pyrefly-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		tempFile = tmp.Name()
		tmp.Close()
		defer os.Remove(tempFile)
		args = append(args, "--output", tempFile, "--output-format", "json")
	} else if outputFile != "" {
		args = append(args, "--output", outputFile)
	}
	if outputFormat == "json" && outputFile == "" {
		args = append(args, "--output-format", "json")
	}

	// Detect config file (pyrefly.toml or pyproject.toml)
	configFiles := []string{"pyrefly.toml", "pyproject.toml"}
	for _, configFile := range configFiles {
		if _, err := os.Stat(filepath.Join(workDirectory, configFile)); err == nil {
			// Pyrefly auto-detects config, so no need to add a flag
			break
		}
	}

	// Add files to check, or "." for current directory
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	cmd := exec.Command(binary, args...)
	cmd.Dir = workDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("failed to run Pyrefly: %w", err)
		}
	}

	if outputFormat == "sarif" {
		jsonOutput, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("failed to read Pyrefly output: %w", err)
		}
		sarifOutput := utils.ConvertPyreflyToSarif(jsonOutput)
		if outputFile != "" {
			err = os.WriteFile(outputFile, sarifOutput, 0644)
			if err != nil {
				return fmt.Errorf("failed to write SARIF output: %w", err)
			}
		} else {
			fmt.Println(string(sarifOutput))
		}
	}
	return nil
}
