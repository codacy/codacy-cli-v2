package cmd

import (
	"codacy/cli-v2/tools"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFileTemplate(t *testing.T) {
	tests := []struct {
		name        string
		tools       []tools.Tool
		expected    []string
		notExpected []string
	}{
		{
			name:  "empty tools list uses defaults",
			tools: []tools.Tool{},
			expected: []string{
				"node@22.2.0",
				"python@3.11.11",
				"eslint@9.3.0",
				"trivy@0.59.1",
				"pylint@3.3.6",
				"pmd@6.55.0",
			},
			notExpected: []string{},
		},
		{
			name: "only eslint enabled",
			tools: []tools.Tool{
				{
					Uuid:    ESLint,
					Name:    "eslint",
					Version: "9.4.0",
				},
			},
			expected: []string{
				"node@22.2.0",
				"eslint@9.4.0",
			},
			notExpected: []string{
				"python@3.11.11",
				"pylint",
				"pmd",
				"trivy",
			},
		},
		{
			name: "only pylint enabled",
			tools: []tools.Tool{
				{
					Uuid:    PyLint,
					Name:    "pylint",
					Version: "3.4.0",
				},
			},
			expected: []string{
				"python@3.11.11",
				"pylint@3.4.0",
			},
			notExpected: []string{
				"node@22.2.0",
				"eslint",
				"pmd",
				"trivy",
			},
		},
		{
			name: "eslint and trivy enabled",
			tools: []tools.Tool{
				{
					Uuid:    ESLint,
					Name:    "eslint",
					Version: "9.4.0",
				},
				{
					Uuid:    Trivy,
					Name:    "trivy",
					Version: "0.60.0",
				},
			},
			expected: []string{
				"node@22.2.0",
				"eslint@9.4.0",
				"trivy@0.60.0",
			},
			notExpected: []string{
				"python@3.11.11",
				"pylint",
				"pmd",
			},
		},
		{
			name: "all tools enabled",
			tools: []tools.Tool{
				{
					Uuid:    ESLint,
					Name:    "eslint",
					Version: "9.4.0",
				},
				{
					Uuid:    Trivy,
					Name:    "trivy",
					Version: "0.60.0",
				},
				{
					Uuid:    PyLint,
					Name:    "pylint",
					Version: "3.4.0",
				},
				{
					Uuid:    PMD,
					Name:    "pmd",
					Version: "6.56.0",
				},
			},
			expected: []string{
				"node@22.2.0",
				"python@3.11.11",
				"eslint@9.4.0",
				"trivy@0.60.0",
				"pylint@3.4.0",
				"pmd@6.56.0",
			},
			notExpected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := configFileTemplate(tt.tools)

			// Check that expected strings are present
			for _, exp := range tt.expected {
				assert.Contains(t, result, exp, "Config file should contain %s", exp)
			}

			// Check that not-expected strings are absent
			for _, notExp := range tt.notExpected {
				assert.NotContains(t, result, notExp, "Config file should not contain %s", notExp)
			}
		})
	}
}

func TestCleanConfigDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some test files in the temp dir
	testFiles := []string{
		"eslint.config.mjs",
		"pylint.rc",
		"ruleset.xml",
		"trivy.yaml",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		assert.NoError(t, err, "Failed to create test file: %s", filePath)
	}

	// Verify files exist
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, len(testFiles), len(files), "Expected %d files before cleaning", len(testFiles))

	// Run the clean function
	err = cleanConfigDirectory(tempDir)
	assert.NoError(t, err, "cleanConfigDirectory should not return an error")

	// Verify all files are gone
	files, err = os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files), "Expected 0 files after cleaning, got %d", len(files))
}
