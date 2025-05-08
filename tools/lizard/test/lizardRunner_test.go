package test

import (
	"codacy/cli-v2/domain"
	"codacy/cli-v2/tools/lizard"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunLizardWithSarifOutput(t *testing.T) {
	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Get the home directory
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Construct the path to the Lizard binary
	globalCache := filepath.Join(homeDirectory, ".cache", "codacy")
	lizardBinary := filepath.Join(globalCache, "tools/lizard@1.17.19/venv/bin/python")

	// Construct the path to the test file
	complexPyPath := filepath.Join(currentDir, "complex.py")

	// Create test patterns
	patterns := []domain.PatternDefinition{
		{
			Id:            "Lizard_nloc-minor",
			Category:      "CodeStyle",
			Level:         "Code",
			SeverityLevel: "Minor",
			Title:         "Method Too Long (Minor)",
			Description:   "Method has too many lines of code",
			Parameters: []domain.ParameterConfiguration{
				{
					Name:    "threshold",
					Value:   "15",
					Default: "15",
				},
			},
		},
		{
			Id:            "Lizard_nloc-medium",
			Category:      "CodeStyle",
			Level:         "Code",
			SeverityLevel: "Medium",
			Title:         "Method Too Long (Medium)",
			Description:   "Method has too many lines of code",
			Parameters: []domain.ParameterConfiguration{
				{
					Name:    "threshold",
					Value:   "25",
					Default: "25",
				},
			},
		},
		{
			Id:            "Lizard_ccn-minor",
			Category:      "CodeStyle",
			Level:         "Code",
			SeverityLevel: "Minor",
			Title:         "Cyclomatic Complexity (Minor)",
			Description:   "Method has high cyclomatic complexity",
			Parameters: []domain.ParameterConfiguration{
				{
					Name:    "threshold",
					Value:   "3",
					Default: "3",
				},
			},
		},
		{
			Id:            "Lizard_ccn-critical",
			Category:      "CodeStyle",
			Level:         "Code",
			SeverityLevel: "Critical",
			Title:         "Cyclomatic Complexity (Critical)",
			Description:   "Method has extremely high cyclomatic complexity",
			Parameters: []domain.ParameterConfiguration{
				{
					Name:    "threshold",
					Value:   "30",
					Default: "30",
				},
			},
		},
	}

	// Create a temporary output file
	outputFile := filepath.Join(currentDir, "output.sarif")
	defer os.Remove(outputFile)

	// Read expected SARIF output
	expectedData, err := os.ReadFile("expected.sarif")
	if err != nil {
		t.Fatalf("Failed to read expected SARIF output: %v", err)
	}
	expectedOutput := strings.TrimSpace(string(expectedData))

	// Run Lizard with SARIF output
	err = lizard.RunLizard(currentDir, lizardBinary, []string{complexPyPath}, outputFile, "sarif", patterns)
	if err != nil {
		t.Fatalf("RunLizard failed: %v", err)
	}

	// Read and parse the SARIF output
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read SARIF output: %v", err)
	}

	// Compare the outputs
	actualOutput := strings.TrimSpace(string(data))
	assert.Equal(t, expectedOutput, actualOutput, "SARIF output does not match expected output")
}
