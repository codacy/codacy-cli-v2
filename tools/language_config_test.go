package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateLanguagesConfigFile(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "codacy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define test tool IDs
	const (
		testESLintID       = "eslint-id"
		testPyLintID       = "pylint-id"
		testTrivyID        = "trivy-id"
		testPMDID          = "pmd-id"
		testDartAnalyzerID = "dartanalyzer-id"
	)

	// Create a map of tool IDs to names
	toolIDMap := map[string]string{
		testESLintID:       "eslint",
		testPyLintID:       "pylint",
		testTrivyID:        "trivy",
		testPMDID:          "pmd",
		testDartAnalyzerID: "dartanalyzer",
	}

	// Create mock tool data
	mockTools := []Tool{
		{
			Uuid: testPyLintID, // Pylint
			Name: "Pylint",
		},
		{
			Uuid: testESLintID, // ESLint
			Name: "ESLint",
		},
	}

	// Call the function under test
	err = CreateLanguagesConfigFile(mockTools, tempDir, toolIDMap)
	if err != nil {
		t.Fatalf("CreateLanguagesConfigFile failed: %v", err)
	}

	// Verify the file was created
	configPath := filepath.Join(tempDir, "languages-config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("languages-config.yaml was not created")
	}

	// Read the file content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Check content has expected tools
	content := string(data)
	if !strings.Contains(content, "name: pylint") {
		t.Errorf("Expected pylint in config, but it was not found")
	}
	if !strings.Contains(content, "name: eslint") {
		t.Errorf("Expected eslint in config, but it was not found")
	}

	// Verify other tools are not included
	if strings.Contains(content, "name: trivy") {
		t.Errorf("Unexpected trivy in config")
	}
	if strings.Contains(content, "name: pmd") {
		t.Errorf("Unexpected pmd in config")
	}

	// Check for flow style arrays
	if !strings.Contains(content, "languages: [") {
		t.Errorf("Expected flow-style array for languages, but not found")
	}
	if !strings.Contains(content, "extensions: [") {
		t.Errorf("Expected flow-style array for extensions, but not found")
	}

	// Test with no tools - should include all tools
	emptyTools := []Tool{}
	err = CreateLanguagesConfigFile(emptyTools, tempDir, toolIDMap)
	if err != nil {
		t.Fatalf("CreateLanguagesConfigFile failed with empty tools: %v", err)
	}

	// Read the file again
	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// With empty tools, all tools should be included
	content = string(data)
	for toolName := range toolIDMap {
		shortName := toolIDMap[toolName]
		if !strings.Contains(content, "name: "+shortName) {
			t.Errorf("Expected %s in config when no tools specified, but it was not found", shortName)
		}
	}
}
