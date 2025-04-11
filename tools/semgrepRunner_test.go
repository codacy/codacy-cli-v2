package tools

import (
	"codacy/cli-v2/plugins"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunSemgrepWithSpecificFiles(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Set up test directories and files
	testDirectory := filepath.Join(currentDirectory, "testdata", "repositories", "semgrep")
	tempResultFile := filepath.Join(os.TempDir(), "semgrep-specific.sarif")
	defer os.Remove(tempResultFile)

	// Create tool info for semgrep
	toolInfo := &plugins.ToolInfo{
		InstallDir: filepath.Join(homeDirectory, ".cache/codacy/tools/semgrep@1.78.0"),
	}

	// Specify files to analyze
	filesToAnalyze := []string{"sample.js"}

	// Run Semgrep analysis on specific files
	err = RunSemgrep(testDirectory, toolInfo, filesToAnalyze, tempResultFile, "sarif")
	if err != nil {
		t.Fatalf("Failed to run semgrep on specific files: %v", err)
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(tempResultFile)
	assert.NoError(t, err, "Failed to stat output file")
	assert.Greater(t, fileInfo.Size(), int64(0), "Output file should not be empty")
}
