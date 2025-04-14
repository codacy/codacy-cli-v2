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
	Tool        Tool         `json:"tool"`
	Results     []Result     `json:"results"`
	Invocations []Invocation `json:"invocations,omitempty"`
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
	EndLine     int `json:"endLine"`
	EndColumn   int `json:"endColumn"`
}

type Invocation struct {
	ExecutionSuccessful bool     `json:"executionSuccessful"`
	ExitCode            int      `json:"exitCode"`
	ExitSignalName      string   `json:"exitSignalName"`
	ExitSignalNumber    int      `json:"exitSignalNumber"`
	Stderr              Artifact `json:"stderr"`
}

type Artifact struct {
	Text string `json:"text"`
}

type MessageText struct {
	Text string `json:"text"`
}

// ReadSarifFile reads a SARIF file and returns its contents
func ReadSarifFile(file string) (SarifReport, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return SarifReport{}, fmt.Errorf("failed to read SARIF file: %w", err)
	}

	var sarif SarifReport
	if err := json.Unmarshal(data, &sarif); err != nil {
		return SarifReport{}, fmt.Errorf("failed to parse SARIF file: %w", err)
	}

	return sarif, nil
}

// WriteSarifFile writes a SARIF report to a file
func WriteSarifFile(sarif SarifReport, outputFile string) error {
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sarif); err != nil {
		return fmt.Errorf("failed to write SARIF: %w", err)
	}

	return nil
}

// AddErrorRun adds an error run to an existing SARIF report
func AddErrorRun(sarif *SarifReport, toolName string, errorMessage string) {
	errorRun := Run{
		Tool: Tool{
			Driver: Driver{
				Name:    toolName,
				Version: "1.0.0",
			},
		},
		Invocations: []Invocation{
			{
				ExecutionSuccessful: false,
				ExitCode:            1,
				ExitSignalName:      "error",
				ExitSignalNumber:    1,
				Stderr: Artifact{
					Text: errorMessage,
				},
			},
		},
		Results: []Result{},
	}
	sarif.Runs = append(sarif.Runs, errorRun)
}

// ConvertPylintToSarif converts Pylint JSON output to SARIF format
func ConvertPylintToSarif(pylintOutput []byte) []byte {
	var issues []PylintIssue
	if err := json.Unmarshal(pylintOutput, &issues); err != nil {
		// If parsing fails, return empty SARIF report with error
		return createEmptySarifReportWithError(err.Error())
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
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true, // Pylint ran successfully if we got here
						ExitCode:            0,
						ExitSignalName:      "",
						ExitSignalNumber:    0,
						Stderr: Artifact{
							Text: "",
						},
					},
				},
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
				Results:     []Result{},
				Invocations: []Invocation{},
			},
		},
	}
	sarifData, err := json.MarshalIndent(emptyReport, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling empty SARIF report: %v\n", err)
		return []byte("{}")
	}
	return sarifData
}

// createEmptySarifReportWithError creates an empty SARIF report with error information
func createEmptySarifReportWithError(errorMessage string) []byte {
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
				Invocations: []Invocation{
					{
						ExecutionSuccessful: false,
						ExitCode:            1,
						ExitSignalName:      "error",
						ExitSignalNumber:    1,
						Stderr: Artifact{
							Text: errorMessage,
						},
					},
				},
			},
		},
	}
	sarifData, err := json.MarshalIndent(emptyReport, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling SARIF report with error: %v\n", err)
		return []byte("{}")
	}
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

		var sarif SimpleSarifReport
		if err := json.Unmarshal(data, &sarif); err != nil {
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
