package tools

import (
	"bufio"
	"bytes"
	"codacy/cli-v2/plugins"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const CodacyApiBase = "https://app.codacy.com"
const codacyToolName = "dartanalyzer"
const patternPrefix = "dartanalyzer_"

func RunDartAnalyzer(workDirectory string, toolInfo *plugins.ToolInfo, files []string, outputFile string, outputFormat string, apiToken string, provider string, owner string, repository string) {

	configFiles := []string{"analysis_options.yaml", "analysis_options.yml"}
	dartAnalyzerPath := filepath.Join(toolInfo.InstallDir, "bin", "dart")
	fmt.Println(dartAnalyzerPath)

	args := []string{"analyze", "--format", "machine"}
	// Add files to analyze - if no files specified, analyze current directory
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	cmd := exec.Command(dartAnalyzerPath, args...)

	cmd.Dir = workDirectory

	fmt.Println("Running", cmd.String())

	// Check if any config file exists
	configExists := false
	for _, configFile := range configFiles {
		if _, err := os.Stat(filepath.Join(workDirectory, configFile)); err == nil {
			configExists = true
			break
		}
	}

	if !configExists {
		fmt.Println("No config file found, using tool defaults")
	} else {
		fmt.Println("Config file found, using it")
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

}

func convertDartAnalyzerOutputToSarif(output string) (string, error) {
	// Create base SARIF structure
	sarif := map[string]interface{}{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		"runs": []map[string]interface{}{
			{
				"results": []map[string]interface{}{},
			},
		},
	}

	// Split output into lines
	lines := strings.Split(output, "\n")

	// Process each line
	for _, line := range lines {
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
		}

		// Add result to SARIF output
		sarif["runs"].([]map[string]interface{})[0]["results"] = append(
			sarif["runs"].([]map[string]interface{})[0]["results"].([]map[string]interface{}),
			result,
		)
	}

	// Convert to JSON
	sarifJson, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling SARIF: %v", err)
	}

	return string(sarifJson), nil
}

func getToolFromCodacy(apiToken string, provider string, owner string, repository string) (*Tool, error) {
	url := fmt.Sprintf("%s/api/v3/analysis/organizations/%s/%s/repositories/%s/tools",
		CodacyApiBase,
		provider,
		owner,
		repository)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-token", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to get tools from Codacy API: %v", resp.Status)
	}

	var response struct {
		Data []Tool `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	for _, tool := range response.Data {
		if tool.Name == codacyToolName {
			return &tool, nil
		}
	}
	return nil, fmt.Errorf("tool %s not found", codacyToolName)
}
