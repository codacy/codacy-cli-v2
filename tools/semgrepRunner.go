package tools

import (
	"codacy/cli-v2/config"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SarifReport represents the structure of a SARIF report
type SarifReport struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []struct {
		Tool struct {
			Driver struct {
				Name  string `json:"name"`
				Rules []any  `json:"rules,omitempty"`
			} `json:"driver"`
		} `json:"tool"`
		Results     []any `json:"results"`
		Invocations []any `json:"invocations,omitempty"`
	} `json:"runs"`
}

// filterRuleDefinitions removes rule definitions from SARIF output
func filterRuleDefinitions(sarifData []byte) ([]byte, error) {
	var report SarifReport
	if err := json.Unmarshal(sarifData, &report); err != nil {
		return nil, fmt.Errorf("failed to parse SARIF data: %w", err)
	}

	// Remove rules from each run
	for i := range report.Runs {
		report.Runs[i].Tool.Driver.Rules = nil
	}

	// Marshal back to JSON with indentation
	return json.MarshalIndent(report, "", "  ")
}

// RunSemgrep executes Semgrep analysis on the specified directory
func RunSemgrep(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	// Construct base command with -m semgrep to run semgrep module
	cmdArgs := []string{"scan"}

	// Defaults from https://github.com/codacy/codacy-semgrep/blob/master/internal/tool/command.go
	cmdArgs = append(cmdArgs, "--max-memory", "2560")
	cmdArgs = append(cmdArgs, "--timeout", "5")
	cmdArgs = append(cmdArgs, "--timeout-threshold", "3")

	cmdArgs = append(cmdArgs, "--disable-version-check")

	// Create a temporary file for SARIF output if needed
	var tempFile string
	if outputFormat == "sarif" {
		tmpFile, err := os.CreateTemp("", "semgrep-*.sarif")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		tempFile = tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tempFile)
		cmdArgs = append(cmdArgs, "--sarif", "--output", tempFile)
	}

	// Define possible Semgrep config file names
	semgrepConfigFiles := []string{"semgrep.yml", "semgrep.yaml", "semgrep/semgrep.yml"}

	// Check if a config file exists in the expected location and use it if present
	if configFile, exists := ConfigFileExists(config.Config, semgrepConfigFiles...); exists {
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

	// Create Semgrep command
	cmd := exec.Command(binary, cmdArgs...)
	cmd.Dir = workDirectory

	if outputFormat != "sarif" && outputFile != "" {
		// If output file is specified and not SARIF, create it and redirect output
		var outputWriter *os.File
		var err error
		outputWriter, err = os.Create(filepath.Clean(outputFile))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.Close()
		cmd.Stdout = outputWriter
	} else if outputFormat != "sarif" {
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

	// If SARIF output was requested, process it
	if outputFormat == "sarif" {
		// Read the temporary SARIF file
		sarifData, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("failed to read SARIF output: %w", err)
		}

		// Filter out rule definitions
		filteredData, err := filterRuleDefinitions(sarifData)
		if err != nil {
			return fmt.Errorf("failed to filter SARIF output: %w", err)
		}

		// Write the filtered output
		if outputFile != "" {
			if err := os.WriteFile(outputFile, filteredData, 0644); err != nil {
				return fmt.Errorf("failed to write filtered SARIF output: %w", err)
			}
		} else {
			fmt.Println(string(filteredData))
		}
	}

	return nil
}
