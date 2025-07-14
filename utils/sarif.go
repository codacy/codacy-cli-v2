package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PylintIssue represents a single issue in Pylint's JSON output
type PylintIssue struct {
	Type      string `json:"type"`
	Module    string `json:"module"`
	Obj       string `json:"obj"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Path      string `json:"path"`
	Symbol    string `json:"symbol"`
	Message   string `json:"message"`
	MessageID string `json:"message-id"`
}

// SarifReport represents the SARIF report structure
type SarifReport struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool    Tool     `json:"tool"`
	Results []Result `json:"results"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	InformationURI string `json:"informationUri"`
	Rules          []Rule `json:"rules"`
}

type Rule struct {
	ID               string         `json:"id"`
	ShortDescription MessageText    `json:"shortDescription"`
	Properties       RuleProperties `json:"properties"`
}

type RuleProperties struct {
	Priority int      `json:"priority,omitempty"`
	Ruleset  string   `json:"ruleset,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	// Add other common properties that might be present
	Precision        string `json:"precision,omitempty"`
	SecuritySeverity string `json:"security-severity,omitempty"`
}

type Result struct {
	RuleID    string      `json:"ruleId"`
	Level     string      `json:"level"`
	Message   MessageText `json:"message"`
	Locations []Location  `json:"locations"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

type Region struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

type MessageText struct {
	Text string `json:"text"`
}

// ConvertPylintToSarif converts Pylint JSON output to SARIF format
func ConvertPylintToSarif(pylintOutput []byte) []byte {
	var issues []PylintIssue
	if err := json.Unmarshal(pylintOutput, &issues); err != nil {
		// If parsing fails, return empty SARIF report
		return createEmptySarifReport()
	}

	// Create SARIF report
	sarifReport := SarifReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:           "Pylint",
						Version:        "3.3.6", // TODO: Get this dynamically
						InformationURI: "https://pylint.org",
					},
				},
				Results: make([]Result, 0, len(issues)),
			},
		},
	}

	// Convert each Pylint issue to SARIF result
	for _, issue := range issues {
		result := Result{
			RuleID: issue.Symbol,
			Level:  getSarifLevel(issue.Type),
			Message: MessageText{
				Text: issue.Message,
			},
			Locations: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{
							URI: issue.Path,
						},
						Region: Region{
							StartLine:   issue.Line,
							StartColumn: issue.Column,
						},
					},
				},
			},
		}
		sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, result)
	}

	sarifData, err := json.MarshalIndent(sarifReport, "", "  ")
	if err != nil {
		return createEmptySarifReport()
	}

	return sarifData
}

// getSarifLevel converts Pylint message type to SARIF level
func getSarifLevel(pylintType string) string {
	switch pylintType {
	case "error", "fatal":
		return "error"
	case "warning":
		return "warning"
	case "convention", "refactor":
		return "note"
	default:
		return "none"
	}
}

// createEmptySarifReport creates an empty SARIF report in case of errors
func createEmptySarifReport() []byte {
	emptyReport := SarifReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:           "Pylint",
						Version:        "3.3.6",
						InformationURI: "https://pylint.org",
					},
				},
				Results: []Result{},
			},
		},
	}
	sarifData, _ := json.MarshalIndent(emptyReport, "", "  ")
	return sarifData
}

type SimpleSarifReport struct {
	Version string            `json:"version"`
	Schema  string            `json:"$schema"`
	Runs    []json.RawMessage `json:"runs"`
}

// MergeSarifOutputs combines multiple SARIF files into a single output file
func MergeSarifOutputs(inputFiles []string, outputFile string) error {
	var mergedSarif SimpleSarifReport
	mergedSarif.Version = "2.1.0"
	mergedSarif.Schema = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"
	mergedSarif.Runs = make([]json.RawMessage, 0)

	for _, file := range inputFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				// Skip if file doesn't exist (tool might have failed)
				continue
			}
			return fmt.Errorf("failed to read SARIF file %s: %w", file, err)
		}

		// Skip empty files
		if len(data) == 0 {
			continue
		}

		var sarif SimpleSarifReport
		if err := json.Unmarshal(data, &sarif); err != nil {
			// If file is empty or invalid JSON, create an empty SARIF report - extra protection from invalid files
			emptySarif := SimpleSarifReport{
				Version: "2.1.0",
				Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
				Runs:    []json.RawMessage{},
			}
			emptyData, err := json.Marshal(emptySarif)
			if err != nil {
				return fmt.Errorf("failed to create empty SARIF report: %w", err)
			}
			if err := json.Unmarshal(emptyData, &sarif); err != nil {
				return fmt.Errorf("failed to parse empty SARIF report: %w", err)
			}
		}

		mergedSarif.Runs = append(mergedSarif.Runs, sarif.Runs...)
	}

	// Create output file
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mergedSarif); err != nil {
		return fmt.Errorf("failed to write merged SARIF: %w", err)
	}

	return nil
}

// FilterRulesFromSarif removes rule definitions from SARIF output if needed
// This should be called separately after MergeSarifOutputs if rule filtering is required
func FilterRulesFromSarif(sarifData []byte) ([]byte, error) {
	// Use a map to preserve all fields during unmarshaling
	var report map[string]interface{}
	if err := json.Unmarshal(sarifData, &report); err != nil {
		return nil, fmt.Errorf("failed to parse SARIF data: %w", err)
	}

	// Navigate to the runs array and remove rules from each run
	if runs, ok := report["runs"].([]interface{}); ok {
		for _, run := range runs {
			if runMap, ok := run.(map[string]interface{}); ok {
				if tool, ok := runMap["tool"].(map[string]interface{}); ok {
					if driver, ok := tool["driver"].(map[string]interface{}); ok {
						// Always set rules to null to maintain consistent output format
						driver["rules"] = nil
					}
				}
			}
		}
	}

	// Marshal back to JSON with indentation
	filteredData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filtered SARIF: %w", err)
	}

	return filteredData, nil
}

// PyreflyIssue represents a single issue in Pyrefly's JSON output.
// Example fields: line, column, stop_line, stop_column, path, code, name, description, concise_description
// See: https://pyrefly.org/en/docs/usage/#output-formats
type PyreflyIssue struct {
	Line               int    `json:"line"`
	Column             int    `json:"column"`
	StopLine           int    `json:"stop_line"`
	StopColumn         int    `json:"stop_column"`
	Path               string `json:"path"`
	Code               int    `json:"code"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	ConciseDescription string `json:"concise_description"`
}

// ConvertPyreflyToSarif converts Pyrefly JSON output to SARIF format
func ConvertPyreflyToSarif(pyreflyOutput []byte) []byte {
	// Pyrefly outputs: { "errors": [ ... ] }
	type pyreflyRoot struct {
		Errors []PyreflyIssue `json:"errors"`
	}
	var root pyreflyRoot
	var sarifReport SarifReport
	cwd, _ := os.Getwd()
	if err := json.Unmarshal(pyreflyOutput, &root); err != nil {
		// If parsing fails, return empty SARIF report with Pyrefly metadata
		sarifReport = SarifReport{
			Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
			Version: "2.1.0",
			Runs: []Run{
				{
					Tool: Tool{
						Driver: Driver{
							Name:           "Pyrefly",
							Version:        "0.22.0",
							InformationURI: "https://pyrefly.org",
						},
					},
					Results: []Result{},
				},
			},
		}
	} else {
		sarifReport = SarifReport{
			Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
			Version: "2.1.0",
			Runs: []Run{
				{
					Tool: Tool{
						Driver: Driver{
							Name:           "Pyrefly",
							Version:        "0.22.0",
							InformationURI: "https://pyrefly.org",
						},
					},
					Results: make([]Result, 0, len(root.Errors)),
				},
			},
		}
		for _, issue := range root.Errors {
			relPath := issue.Path
			if rel, err := filepath.Rel(cwd, issue.Path); err == nil {
				relPath = rel
			}
			result := Result{
				RuleID: issue.Name,
				Level:  "error", // Pyrefly only reports errors
				Message: MessageText{
					Text: issue.Description,
				},
				Locations: []Location{
					{
						PhysicalLocation: PhysicalLocation{
							ArtifactLocation: ArtifactLocation{
								URI: relPath,
							},
							Region: Region{
								StartLine:   issue.Line,
								StartColumn: issue.Column,
							},
						},
					},
				},
			}
			sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, result)
		}
	}
	sarifData, err := json.MarshalIndent(sarifReport, "", "  ")
	if err != nil {
		// If marshaling fails, return a minimal SARIF report with Pyrefly metadata
		return []byte(`{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Pyrefly",
          "version": "0.22.0",
          "informationUri": "https://pyrefly.org"
        }
      },
      "results": []
    }
  ]
}`)
	}
	return sarifData
}
