package tools

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/utils"
	"fmt"
	"os"
	"os/exec"
)

func RunPylint(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {

	// Construct base command with -m pylint to run pylint module
	args := []string{"-m", "pylint"}

	// Always use JSON output format since we'll convert to SARIF if needed
	args = append(args, "--output-format=json")

	// Check if a config file exists in the expected location and use it if present
	if configFile, exists := ConfigFileExists(config.Config, "pylint.rc"); exists {
		args = append(args, fmt.Sprintf("--rcfile=%s", configFile))
	}

	// Create a temporary file for JSON output if we need to convert to SARIF
	var tempFile string
	if outputFormat == "sarif" {
		tmp, err := os.CreateTemp("", "pylint-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		tempFile = tmp.Name()
		tmp.Close()
		defer os.Remove(tempFile)
		args = append(args, fmt.Sprintf("--output=%s", tempFile))
	} else if outputFile != "" {
		args = append(args, fmt.Sprintf("--output=%s", outputFile))
	}

	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	// Create and run command
	cmd := exec.Command(binary, args...)
	cmd.Dir = workDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Pylint returns non-zero exit code when it finds issues
		// We should not treat this as an error
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("failed to run Pylint: %w", err)
		}
	}

	// If SARIF output is requested, convert JSON to SARIF
	if outputFormat == "sarif" {
		jsonOutput, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("failed to read pylint output: %w", err)
		}

		sarifOutput := utils.ConvertPylintToSarif(jsonOutput)

		if outputFile != "" {
			err = os.WriteFile(outputFile, sarifOutput, constants.DefaultFilePerms)
			if err != nil {
				return fmt.Errorf("failed to write SARIF output: %w", err)
			}
		} else {
			fmt.Println(string(sarifOutput))
		}
	}

	return nil
}
