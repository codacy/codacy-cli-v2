package lizard

import (
	"bytes"
	"codacy/cli-v2/domain"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// RunLizard runs the Lizard tool and returns any issues found
func RunLizard(workDirectory string, binary string, files []string, outputFile string, outputFormat string, patterns []domain.PatternDefinition) error {

	if len(patterns) == 0 {
		return fmt.Errorf("no valid patterns found in configuration")
	}

	// Construct base command with lizard module
	args := []string{"-m", "lizard", "-V"}

	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	// For non-SARIF output, let Lizard handle file output directly
	if outputFormat != "sarif" && outputFile != "" {
		args = append(args, "-o", outputFile)
	}

	// Run the command
	cmd := exec.Command(binary, args...)
	cmd.Dir = workDirectory
	cmd.Stderr = os.Stderr

	var err error
	// For SARIF output, we need to capture and parse the output
	if outputFormat == "sarif" {
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err = cmd.Run()
		if err != nil {
			// Lizard returns 1 when it finds issues, which is not a failure
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// Parse the output and generate issues
				results, parseErr := parseLizardResults(stdout.String())
				if parseErr != nil {
					return fmt.Errorf("failed to parse Lizard output: %w", parseErr)
				}

				issues := generateIssuesFromResults(results, patterns)

				// Convert issues to SARIF Report
				sarifReport := convertIssuesToSarif(issues, patterns)

				// Marshal SARIF Report report to Sarif
				sarifData, err := json.MarshalIndent(sarifReport, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal SARIF report: %w", err)
				}

				// Write SARIF output to file if specified, else stdout
				if outputFile != "" {
					err = os.WriteFile(outputFile, sarifData, 0644)
					if err != nil {
						return fmt.Errorf("failed to write SARIF output: %w", err)
					}
				} else {
					fmt.Println(string(sarifData))
				}

				return nil
			}
			return fmt.Errorf("failed to run Lizard: %w", err)
		}
	} else {
		// For non-SARIF output, let Lizard handle stdout
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			// Lizard returns 1 when it finds issues, which is not a failure
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return nil
			}
			return fmt.Errorf("failed to run Lizard: %w", err)
		}
	}

	return nil
}
