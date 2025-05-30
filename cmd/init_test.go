package cmd

import (
	"codacy/cli-v2/cmd/configsetup"
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
				"eslint@9.3.0",
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
					Uuid:    configsetup.ESLint,
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
					Uuid:    configsetup.PyLint,
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
					Uuid:    configsetup.ESLint,
					Name:    "eslint",
					Version: "9.4.0",
				},
				{
					Uuid:    configsetup.Trivy,
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
					Uuid:    configsetup.ESLint,
					Name:    "eslint",
					Version: "9.4.0",
				},
				{
					Uuid:    configsetup.Trivy,
					Name:    "trivy",
					Version: "0.60.0",
				},
				{
					Uuid:    configsetup.PyLint,
					Name:    "pylint",
					Version: "3.4.0",
				},
				{
					Uuid:    configsetup.PMD,
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
			result := configsetup.ConfigFileTemplate(tt.tools)

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
	err = configsetup.CleanConfigDirectory(tempDir)
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
	rulesPath := filepath.Join("..", "plugins", "tools", "semgrep", "rules.yaml")
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		t.Skipf("plugins/tools/semgrep/rules.yaml not found at %s; skipping test", rulesPath)
	}

	// Change to the temp directory to simulate a new project
	err = os.Chdir(tempDir)
	assert.NoError(t, err, "Failed to change working directory to tempDir")

	// Simulate running init with no token
	currentInitFlags := domain.InitFlags{
		ApiToken:     "",
		Provider:     "",
		Organization: "",
		Repository:   "",
	}

	// Call the Run logic from initCmd
	if err := config.Config.CreateLocalCodacyDir(); err != nil {
		t.Fatalf("Failed to create local codacy directory: %v", err)
	}

	toolsConfigDir := config.Config.ToolsConfigDirectory()
	if err := os.MkdirAll(toolsConfigDir, utils.DefaultDirPerms); err != nil {
		t.Fatalf("Failed to create tools-configs directory: %v", err)
	}

	cliLocalMode := len(currentInitFlags.ApiToken) == 0
	if cliLocalMode {
		noTools := []domain.Tool{}
		err := configsetup.CreateConfigurationFiles(noTools, cliLocalMode)
		assert.NoError(t, err, "CreateConfigurationFiles should not return an error")
		if err := configsetup.BuildDefaultConfigurationFiles(toolsConfigDir, currentInitFlags); err != nil {
			t.Fatalf("Failed to build default configuration files: %v", err)
		}
		if err := configsetup.CreateLanguagesConfigFileLocal(toolsConfigDir); err != nil {
			t.Fatalf("Failed to create languages config file: %v", err)
		}
	}

	// Assert that the expected config files are created
	codacyDir := config.Config.LocalCodacyDirectory()
	expectedFiles := []string{
		filepath.Join("tools-configs", "eslint.config.mjs"),
		filepath.Join("tools-configs", "trivy.yaml"),
		filepath.Join("tools-configs", "ruleset.xml"),
		filepath.Join("tools-configs", "pylint.rc"),
		filepath.Join("tools-configs", "analysis_options.yaml"),
		filepath.Join("tools-configs", "semgrep.yaml"),
		filepath.Join("tools-configs", "lizard.yaml"),
		"codacy.yaml",
		"cli-config.yaml",
		filepath.Join("tools-configs", "languages-config.yaml"),
		".gitignore",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(codacyDir, file)
		if file == ".gitignore" {
			filePath = filepath.Join(config.Config.LocalCodacyDirectory(), file)
		}

		_, err := os.Stat(filePath)
		assert.NoError(t, err, "Expected config file %s to be created at %s", file, filePath)
	}
}
