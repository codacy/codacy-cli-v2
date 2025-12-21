package test

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/tools/lizard"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunLizardWithSarifOutput(t *testing.T) {
	config.Init()

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
	lizardBinary := filepath.Join(globalCache, "tools/lizard@1.17.31/venv/bin/python")

	// Construct the path to the test file
	complexPyPath := filepath.Join(currentDir, "complex.py")

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
	err = lizard.RunLizard(currentDir, lizardBinary, []string{complexPyPath}, outputFile, "sarif")
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
