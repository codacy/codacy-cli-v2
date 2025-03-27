package tools

import (
	"fmt"
	"os"
	"os/exec"
)

// RunTrivy executes Trivy vulnerability scanner with the specified options
func RunTrivy(repositoryToAnalyseDirectory string, trivyBinary string, pathsToCheck []string, outputFile string, outputFormat string) error {
	cmd := exec.Command(trivyBinary, "fs")

	// Add format options
	if outputFile != "" {
		// When writing to file, use SARIF format
		cmd.Args = append(cmd.Args, "--format", "sarif", "--output", outputFile)
	} else if outputFormat == "sarif" {
		// When outputting to terminal in SARIF format
		cmd.Args = append(cmd.Args, "--format", "sarif")
	}

	// Add severity filtering to match common expectations
	// cmd.Args = append(cmd.Args, "--severity", "HIGH,CRITICAL")

	// Add specific targets or use current directory
	if len(pathsToCheck) > 0 {
		cmd.Args = append(cmd.Args, pathsToCheck...)
	} else {
		cmd.Args = append(cmd.Args, ".")
	}

	// Set working directory
	cmd.Dir = repositoryToAnalyseDirectory
	cmd.Stderr = os.Stderr

	// If outputting to terminal and not in SARIF format, direct to stdout
	if outputFile == "" && outputFormat != "sarif" {
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}

	// If outputting SARIF to terminal, capture output and print
	if outputFile == "" && outputFormat == "sarif" {
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("trivy scan failed: %w", err)
		}

		fmt.Println(string(output))
		return nil
	}

	// If outputting to file, just run the command
	return cmd.Run()
}
