package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const patternPrefix = "dartanalyzer_"

func RunDartAnalyzer(workDirectory string, installationDirectory string, binary string, files []string, outputFile string, outputFormat string) error {

	configFiles := []string{"analysis_options.yaml", "analysis_options.yml"}
	dartAnalyzerPath := filepath.Join(installationDirectory, "bin", "dart")

	args := []string{"analyze", "--format", "machine"}
	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	cmd := exec.Command(dartAnalyzerPath, args...)

	cmd.Dir = workDirectory

	// Check if any config file exists
	configExists := false
	for _, configFile := range configFiles {
		if _, err := os.Stat(filepath.Join(workDirectory, configFile)); err == nil {
			configExists = true
			break
		}
	}

	if !configExists {
		log.Println("No config file found, using tool defaults")
	} else {
		log.Println("Config file found, using it")
	}

	// For SARIF output, we need to capture the output and transform it
	if outputFormat == "sarif" {
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		cmd.Run()

		// Convert Dart Analyzer output to SARIF format
		sarif := map[string]interface{}{
			"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
			"runs": []map[string]interface{}{
				{
					"tool": map[string]interface{}{
						"driver": map[string]interface{}{
							"name": "dartanalyzer",
						},
					},
					"results": []map[string]interface{}{},
				},
			},
		}

		// Parse Dart Analyzer output and convert to SARIF
		// Format is typically: file:line:col: severity: message
		scanner := bufio.NewScanner(strings.NewReader(stdout.String()))
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// Split line into fields
			fields := strings.Split(line, "|")
			if len(fields) < 8 {
				continue
			}

			// Extract fields
			file := fields[3]
			lineNum, _ := strconv.Atoi(fields[4])
			message := fields[7]
			ruleId := fields[2]

			// Create result object
			result := map[string]interface{}{
				"message": map[string]string{
					"text": message,
				},
				"locations": []map[string]interface{}{
					{
						"physicalLocation": map[string]interface{}{
							"artifactLocation": map[string]interface{}{
								"uri": file,
							},
							"region": map[string]interface{}{
								"startLine": lineNum,
							},
						},
					},
				},
				"ruleId": ruleId,
			}

			// Add result to SARIF output
			sarif["runs"].([]map[string]interface{})[0]["results"] = append(
				sarif["runs"].([]map[string]interface{})[0]["results"].([]map[string]interface{}),
				result,
			)
		}

		// Write SARIF output to file if specified
		if outputFile != "" {
			sarifJson, _ := json.MarshalIndent(sarif, "", "  ")
			err := os.WriteFile(outputFile, sarifJson, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing SARIF output: %v\n", err)
			}
		}
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	return nil
}
