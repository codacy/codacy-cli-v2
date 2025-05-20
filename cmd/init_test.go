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
		{
			name: "all tools enabled",
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

func TestInitCommand_NoToken(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	defer os.Chdir(originalWD)

	// Use the real plugins/tools/semgrep/rules.yaml file
	rulesPath := filepath.Join("plugins", "tools", "semgrep", "rules.yaml")
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		t.Skip("plugins/tools/semgrep/rules.yaml not found; skipping test")
	}

	// Change to the temp directory to simulate a new project
	err = os.Chdir(tempDir)
	assert.NoError(t, err, "Failed to change working directory to tempDir")

	// Simulate running init with no token
	initFlags.ApiToken = ""
	initFlags.Provider = ""
	initFlags.Organization = ""
	initFlags.Repository = ""

	// Call the Run logic from initCmd
	if err := config.Config.CreateLocalCodacyDir(); err != nil {
		t.Fatalf("Failed to create local codacy directory: %v", err)
	}

	toolsConfigDir := config.Config.ToolsConfigDirectory()
	if err := os.MkdirAll(toolsConfigDir, utils.DefaultFilePerms); err != nil {
		t.Fatalf("Failed to create tools-configs directory: %v", err)
	}

	cliLocalMode := len(initFlags.ApiToken) == 0
	if cliLocalMode {
		noTools := []domain.Tool{}
		err := createConfigurationFiles(noTools, cliLocalMode)
		assert.NoError(t, err, "createConfigurationFiles should not return an error")
		if err := buildDefaultConfigurationFiles(toolsConfigDir); err != nil {
			t.Fatalf("Failed to build default configuration files: %v", err)
		}
	}

	// Assert that the expected config files are created
	codacyDir := config.Config.LocalCodacyDirectory()
	expectedFiles := []string{
		"tools-configs/eslint.config.mjs",
		"tools-configs/trivy.yaml",
		"tools-configs/ruleset.xml",
		"tools-configs/pylint.rc",
		"tools-configs/analysis_options.yaml",
		"tools-configs/semgrep.yaml",
		"tools-configs/lizard.yaml",
		"codacy.yaml",
		"cli-config.yaml",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(codacyDir, file)
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "Expected config file %s to be created", file)
	}
}
