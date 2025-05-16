package utils

import (
	"encoding/json"
	"fmt"
	"os"
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
	Version string `json:"version"`
	Schema  string `json:"$schema"`
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

		// Filter out rule definitions from each input file
		filteredData, err := FilterRuleDefinitions(data)
		if err != nil {
			return fmt.Errorf("failed to filter rules from SARIF file %s: %w", file, err)
		}

		var sarif SimpleSarifReport
		if err := json.Unmarshal(filteredData, &sarif); err != nil {
			return fmt.Errorf("failed to parse SARIF file %s: %w", file, err)
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

// FilterRuleDefinitions removes rule definitions from SARIF output
func FilterRuleDefinitions(sarifData []byte) ([]byte, error) {
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
