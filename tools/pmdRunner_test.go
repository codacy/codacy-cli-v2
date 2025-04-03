package tools

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SarifResult represents a single result in a SARIF report
type SarifResult struct {
	RuleID  string `json:"ruleId"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []struct {
		PhysicalLocation struct {
			ArtifactLocation struct {
				URI string `json:"uri"`
			} `json:"artifactLocation"`
			Region struct {
				StartLine   int `json:"startLine"`
				StartColumn int `json:"startColumn"`
				EndLine     int `json:"endLine"`
				EndColumn   int `json:"endColumn"`
			} `json:"region"`
		} `json:"physicalLocation"`
	} `json:"locations"`
}

// SarifReport represents the structure of a SARIF report
type SarifReport struct {
	Runs []struct {
		Results []SarifResult `json:"results"`
	} `json:"runs"`
}

func TestRunPmdToFile(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Use the correct path relative to tools directory
	testDirectory := filepath.Join(currentDirectory, "testdata", "repositories", "pmd")
	tempResultFile := filepath.Join(os.TempDir(), "pmd.sarif")
	defer os.Remove(tempResultFile)

	// Use absolute paths
	repositoryToAnalyze := testDirectory
	// Use the standard ruleset file for testing the PMD runner functionality
	rulesetFile := filepath.Join(testDirectory, "pmd-ruleset.xml")

	// Use the same path as defined in plugin.yaml
	pmdBinary := filepath.Join(homeDirectory, ".cache/codacy/tools/pmd@6.55.0/pmd-bin-6.55.0/bin/run.sh")

	// Run PMD
	err = RunPmd(repositoryToAnalyze, pmdBinary, nil, tempResultFile, "sarif", rulesetFile)
	if err != nil {
		t.Fatalf("Failed to run pmd: %v", err)
	}

	// Check if the output file was created
	obtainedSarifBytes, err := os.ReadFile(tempResultFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Normalize paths in the obtained SARIF output
	obtainedSarif := string(obtainedSarifBytes)
	obtainedSarif = strings.ReplaceAll(obtainedSarif, currentDirectory+"/", "")

	// Parse the normalized SARIF output
	var sarifReport SarifReport
	err = json.Unmarshal([]byte(obtainedSarif), &sarifReport)
	if err != nil {
		t.Fatalf("Failed to parse SARIF output: %v", err)
	}

	// Verify we have results
	assert.NotEmpty(t, sarifReport.Runs, "SARIF report should have at least one run")
	assert.NotEmpty(t, sarifReport.Runs[0].Results, "SARIF report should have at least one result")

	// Define expected violations
	expectedViolations := map[string]bool{
		"UnusedPrivateField":    false,
		"ShortVariable":         false,
		"AtLeastOneConstructor": false,
		"CommentRequired":       false,
	}

	// Check each result
	for _, result := range sarifReport.Runs[0].Results {
		// Mark this rule as found
		expectedViolations[result.RuleID] = true

		// Verify the file path is correct
		assert.Contains(t, result.Locations[0].PhysicalLocation.ArtifactLocation.URI, "RulesBreaker.java",
			"Violation should be in RulesBreaker.java")

		// Verify line numbers are reasonable
		assert.Greater(t, result.Locations[0].PhysicalLocation.Region.StartLine, 0,
			"Start line should be positive")
		assert.Less(t, result.Locations[0].PhysicalLocation.Region.StartLine, 30,
			"Start line should be within the file")
	}

	// Verify all expected violations were found
	for ruleID, found := range expectedViolations {
		assert.True(t, found, "Expected violation %s was not found", ruleID)
	}
}
