package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFileTemplate(t *testing.T) {
	tests := []struct {
		name        string
		tools       []domain.Tool
		expected    []string
		notExpected []string
	}{
		{
			name:  "empty tools list uses defaults",
			tools: []domain.Tool{},
			expected: []string{
				"node@22.2.0",
				"python@3.11.11",
				"eslint@8.57.0",
				"trivy@0.59.1",
				"pylint@3.3.6",
				"pmd@6.55.0",
			},
			notExpected: []string{},
		},
		{
			name: "only eslint enabled",
			tools: []domain.Tool{
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
			tools: []domain.Tool{
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
			tools: []domain.Tool{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := configFileTemplate(tt.tools)

			for _, exp := range tt.expected {
				assert.Contains(t, result, exp)
			}

			for _, notExp := range tt.notExpected {
				assert.NotContains(t, result, notExp)
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
		err := os.WriteFile(filePath, []byte("test content"), utils.DefaultFilePerms)
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

func TestInitCommand_LanguageDetection(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	defer os.Chdir(originalWD)

	// Create test files for different languages
	testFiles := map[string]string{
		"src/main.go":         "package main",
		"src/app.js":          "console.log('hello')",
		"src/lib.py":          "print('hello')",
		"src/Main.java":       "class Main {}",
		"src/styles.css":      "body { margin: 0; }",
		"src/config.json":     "{}",
		"src/Dockerfile":      "FROM ubuntu",
		"src/app.dart":        "void main() {}",
		"src/test.rs":         "fn main() {}",
		"vendor/ignore.js":    "// should be ignored",
		"node_modules/pkg.js": "// should be ignored",
		".git/config":         "// should be ignored",
	}

	// Create the files in the temporary directory
	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Change to the temp directory
	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Create necessary directories
	err = config.Config.CreateLocalCodacyDir()
	assert.NoError(t, err)
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	err = os.MkdirAll(toolsConfigDir, utils.DefaultFilePerms)
	assert.NoError(t, err)

	// Reset initFlags to simulate local mode
	initFlags.ApiToken = ""
	initFlags.Provider = ""
	initFlags.Organization = ""
	initFlags.Repository = ""

	// Run the init command
	initCmd.Run(initCmd, []string{})

	// Verify that the configuration files were created
	codacyYaml := filepath.Join(config.Config.LocalCodacyDirectory(), "codacy.yaml")
	assert.FileExists(t, codacyYaml)

	// Read and verify the codacy.yaml content
	content, err := os.ReadFile(codacyYaml)
	assert.NoError(t, err)
	contentStr := string(content)

	// Check that appropriate tools are enabled based on detected languages
	assert.Contains(t, contentStr, "eslint@", "ESLint should be enabled for JavaScript")
	assert.Contains(t, contentStr, "pylint@", "PyLint should be enabled for Python")
	assert.Contains(t, contentStr, "semgrep@", "Semgrep should be enabled for Go and Rust")
	assert.Contains(t, contentStr, "dartanalyzer@", "DartAnalyzer should be enabled for Dart")
	assert.Contains(t, contentStr, "trivy@", "Trivy should always be enabled")
	assert.Contains(t, contentStr, "lizard@", "Lizard should be enabled when supported languages are detected")

	// Verify that tool configuration files were created
	expectedConfigFiles := []string{
		"eslint.config.mjs",
		"pylint.rc",
		"trivy.yaml",
		"semgrep.yaml",
		"analysis_options.yaml", // for dartanalyzer
		"lizard.yaml",
	}

	for _, configFile := range expectedConfigFiles {
		assert.FileExists(t, filepath.Join(toolsConfigDir, configFile))
	}
}

func TestInitCommand_NoLanguagesDetected(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	defer os.Chdir(originalWD)

	// Change to the temp directory (empty, no source files)
	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Create necessary directories
	err = config.Config.CreateLocalCodacyDir()
	assert.NoError(t, err)
	toolsConfigDir := config.Config.ToolsConfigDirectory()
	err = os.MkdirAll(toolsConfigDir, utils.DefaultFilePerms)
	assert.NoError(t, err)

	// Reset initFlags to simulate local mode
	initFlags.ApiToken = ""
	initFlags.Provider = ""
	initFlags.Organization = ""
	initFlags.Repository = ""

	// Run the init command
	initCmd.Run(initCmd, []string{})

	// Verify that the configuration files were created
	codacyYaml := filepath.Join(config.Config.LocalCodacyDirectory(), "codacy.yaml")
	assert.FileExists(t, codacyYaml)

	// Read and verify the codacy.yaml content
	content, err := os.ReadFile(codacyYaml)
	assert.NoError(t, err)
	contentStr := string(content)

	// Check that only Trivy is enabled when no languages are detected
	assert.Contains(t, contentStr, "trivy@", "Trivy should always be enabled")
	assert.NotContains(t, contentStr, "eslint@", "ESLint should not be enabled")
	assert.NotContains(t, contentStr, "pylint@", "PyLint should not be enabled")
	assert.NotContains(t, contentStr, "semgrep@", "Semgrep should not be enabled")
	assert.NotContains(t, contentStr, "dartanalyzer@", "DartAnalyzer should not be enabled")
	assert.NotContains(t, contentStr, "lizard@", "Lizard should not be enabled")

	// Verify that only Trivy configuration file was created
	assert.FileExists(t, filepath.Join(toolsConfigDir, "trivy.yaml"))
}

func TestToolNameFromUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected string
	}{
		{
			name:     "ESLint",
			uuid:     ESLint,
			expected: "eslint",
		},
		{
			name:     "Trivy",
			uuid:     Trivy,
			expected: "trivy",
		},
		{
			name:     "PyLint",
			uuid:     PyLint,
			expected: "pylint",
		},
		{
			name:     "PMD",
			uuid:     PMD,
			expected: "pmd",
		},
		{
			name:     "DartAnalyzer",
			uuid:     DartAnalyzer,
			expected: "dartanalyzer",
		},
		{
			name:     "Semgrep",
			uuid:     Semgrep,
			expected: "semgrep",
		},
		{
			name:     "Lizard",
			uuid:     Lizard,
			expected: "lizard",
		},
		{
			name:     "Unknown UUID",
			uuid:     "unknown-uuid",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toolNameFromUUID(tt.uuid)
			assert.Equal(t, tt.expected, result)
		})
	}
}
