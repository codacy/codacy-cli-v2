package tools

import (
	"codacy/cli-v2/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunLicenseSim runs the license-sim tool (duncan.py) in the Python venv with the required environment variable.
func RunLicenseSim(workDirectory string, binary string, files []string, outputFile string, outputFormat string) error {
	var fileToCheck string
	var ext string
	if len(files) > 0 {
		fileToCheck = files[0]
		ext = filepath.Ext(fileToCheck)
		if ext != "" {
			ext = ext[1:] // remove dot
			// Hardcode support for .php files
			if ext == ".php" {
				ext = "php"
			}
		} else {
			ext = "py" // default
		}
	} else {
		return fmt.Errorf("No file specified for license-sim")
	}

	parts := strings.Split(binary, " ")
	cmdArgs := append(parts[1:], "search", "-f", fileToCheck, "-e", ext)

	if outputFormat == "sarif" {
		tempFile, err := os.CreateTemp("", "license-sim-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		tempFilePath := tempFile.Name()
		tempFile.Close()
		defer os.Remove(tempFilePath)

		cmdArgs = append(cmdArgs, "--json")
		cmd := exec.Command(parts[0], cmdArgs...)
		cmd.Dir = workDirectory
		cmd.Env = append(os.Environ(), "KMP_DUPLICATE_LIB_OK=TRUE")
		outFile, err := os.Create(tempFilePath)
		if err != nil {
			return fmt.Errorf("failed to redirect output: %w", err)
		}
		defer outFile.Close()
		cmd.Stdout = outFile
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run license-sim: %w", err)
		}

		jsonOutput, err := os.ReadFile(tempFilePath)
		if err != nil {
			return fmt.Errorf("failed to read license-sim output: %w", err)
		}

		sarifOutput := utils.ConvertLicenseSimToSarifWithFile(jsonOutput, fileToCheck)

		if outputFile != "" {
			err = os.WriteFile(outputFile, sarifOutput, 0644)
			if err != nil {
				return fmt.Errorf("failed to write SARIF output: %w", err)
			}
		} else {
			fmt.Println(string(sarifOutput))
		}
		return nil
	}

	// Non-SARIF output
	cmd := exec.Command(parts[0], cmdArgs...)
	cmd.Dir = workDirectory
	cmd.Env = append(os.Environ(), "KMP_DUPLICATE_LIB_OK=TRUE")

	if outputFile != "" {
		outputWriter, err := os.Create(filepath.Clean(outputFile))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.Close()
		cmd.Stdout = outputWriter
		cmd.Stderr = outputWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run license-sim: %w", err)
	}
	return nil
}
