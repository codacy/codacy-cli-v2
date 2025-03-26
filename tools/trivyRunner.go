package tools

import (
	"encoding/json"
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
		for _, path := range pathsToCheck {
			cmd.Args = append(cmd.Args, path)
		}
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

// TrivyJsonToSarif converts Trivy JSON output to SARIF format
// Note: This function is not needed when using Trivy's built-in SARIF output
// but is included for completeness if you need custom conversions
func TrivyJsonToSarif(trivyJsonFile string, sarifOutputFile string) error {
	// Read Trivy JSON output
	data, err := os.ReadFile(trivyJsonFile)
	if err != nil {
		return fmt.Errorf("failed to read Trivy JSON file: %w", err)
	}

	// Parse Trivy JSON
	var trivyResults map[string]interface{}
	if err := json.Unmarshal(data, &trivyResults); err != nil {
		return fmt.Errorf("failed to parse Trivy JSON: %w", err)
	}

	// Build SARIF structure
	sarif := map[string]interface{}{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":            "Trivy",
						"informationUri":  "https://github.com/aquasecurity/trivy",
						"semanticVersion": "1.0.0",
						"rules":           []interface{}{},
					},
				},
				"results": []interface{}{},
			},
		},
	}

	// Convert results and rules (simplified implementation)
	// In a real implementation, you would iterate through Trivy results
	// and convert each one to SARIF format

	// Write SARIF output
	sarifData, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SARIF data: %w", err)
	}

	if err := os.WriteFile(sarifOutputFile, sarifData, 0644); err != nil {
		return fmt.Errorf("failed to write SARIF file: %w", err)
	}

	return nil
}
