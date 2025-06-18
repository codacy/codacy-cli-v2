package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codacy/cli-v2/config"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckIfConfigExistsAndIsNeededBehavior tests the behavior of checkIfConfigExistsAndIsNeeded
// without creating actual config files in the project directory
func TestCheckIfConfigExistsAndIsNeededBehavior(t *testing.T) {
	// Save original state
	originalFlags := initFlags
	originalConfig := config.Config
	originalWorkingDir, _ := os.Getwd()
	defer func() {
		initFlags = originalFlags
		config.Config = originalConfig
		os.Chdir(originalWorkingDir)
	}()

	tests := []struct {
		name          string
		toolName      string
		cliLocalMode  bool
		apiToken      string
		description   string
		expectNoError bool
	}{
		{
			name:          "tool_without_config_file",
			toolName:      "unsupported-tool",
			cliLocalMode:  false,
			apiToken:      "test-token",
			description:   "Tool that doesn't use config files should return without error",
			expectNoError: true,
		},
		{
			name:          "eslint_local_mode",
			toolName:      "eslint",
			cliLocalMode:  true,
			apiToken:      "",
			description:   "ESLint in local mode should not error",
			expectNoError: true,
		},
		{
			name:          "eslint_remote_mode_without_token",
			toolName:      "eslint",
			cliLocalMode:  false,
			apiToken:      "",
			description:   "ESLint in remote mode without token should not error",
			expectNoError: true,
		},
		{
			name:          "trivy_local_mode",
			toolName:      "trivy",
			cliLocalMode:  true,
			apiToken:      "",
			description:   "Trivy in local mode should not error",
			expectNoError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and change to it to avoid creating files in project dir
			tmpDir, err := os.MkdirTemp("", "codacy-test-isolated-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Change to temp directory to avoid creating config files in project
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Setup initFlags
			initFlags = domain.InitFlags{
				ApiToken: tt.apiToken,
			}

			// Mock config to use our temporary directory
			config.Config = *config.NewConfigType(tmpDir, tmpDir, tmpDir)

			// Execute the function - this tests it doesn't panic or return unexpected errors
			err = checkIfConfigExistsAndIsNeeded(tt.toolName, tt.cliLocalMode)

			// Verify results
			if tt.expectNoError {
				assert.NoError(t, err, "Function should not return error: %s", tt.description)
			} else {
				assert.Error(t, err, "Function should return error: %s", tt.description)
			}
		})
	}
}

// TestToolConfigFileNameMapCompleteness ensures all expected tools have config mappings
func TestToolConfigFileNameMapCompleteness(t *testing.T) {
	expectedTools := map[string]string{
		"eslint":       constants.ESLintConfigFileName,
		"trivy":        constants.TrivyConfigFileName,
		"pmd":          constants.PMDConfigFileName,
		"pylint":       constants.PylintConfigFileName,
		"dartanalyzer": constants.DartAnalyzerConfigFileName,
		"semgrep":      constants.SemgrepConfigFileName,
		"revive":       constants.ReviveConfigFileName,
		"lizard":       constants.LizardConfigFileName,
	}

	t.Run("all_expected_tools_present", func(t *testing.T) {
		for toolName, expectedFileName := range expectedTools {
			actualFileName, exists := constants.ToolConfigFileNames[toolName]
			assert.True(t, exists, "Tool %s should exist in constants.ToolConfigFileNames map", toolName)
			assert.Equal(t, expectedFileName, actualFileName, "Config filename for %s should match expected", toolName)
		}
	})

	t.Run("no_unexpected_tools", func(t *testing.T) {
		for toolName := range constants.ToolConfigFileNames {
			_, expected := expectedTools[toolName]
			assert.True(t, expected, "Unexpected tool %s found in constants.ToolConfigFileNames map", toolName)
		}
	})

	t.Run("config_files_have_proper_extensions", func(t *testing.T) {
		validExtensions := map[string]bool{
			".mjs":  true,
			".js":   true,
			".yaml": true,
			".yml":  true,
			".xml":  true,
			".rc":   true,
			".toml": true,
		}

		for toolName, fileName := range constants.ToolConfigFileNames {
			ext := filepath.Ext(fileName)
			assert.True(t, validExtensions[ext], "Tool %s has config file %s with unexpected extension %s", toolName, fileName, ext)
		}
	})
}

// TestCheckIfConfigExistsAndIsNeededEdgeCases tests edge cases and error conditions
func TestCheckIfConfigExistsAndIsNeededEdgeCases(t *testing.T) {
	originalFlags := initFlags
	originalConfig := config.Config
	originalWorkingDir, _ := os.Getwd()
	defer func() {
		initFlags = originalFlags
		config.Config = originalConfig
		os.Chdir(originalWorkingDir)
	}()

	// Create temporary directory for edge case tests
	tmpDir, err := os.MkdirTemp("", "codacy-test-edge-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Mock config
	config.Config = *config.NewConfigType(tmpDir, tmpDir, tmpDir)

	t.Run("empty_tool_name", func(t *testing.T) {
		err := checkIfConfigExistsAndIsNeeded("", false)
		assert.NoError(t, err, "Empty tool name should not cause error")
	})

	t.Run("tool_name_with_special_characters", func(t *testing.T) {
		err := checkIfConfigExistsAndIsNeeded("tool-with-dashes_and_underscores", false)
		assert.NoError(t, err, "Tool name with special characters should not cause error")
	})

	t.Run("very_long_tool_name", func(t *testing.T) {
		longToolName := strings.Repeat("verylongtoolname", 10)
		err := checkIfConfigExistsAndIsNeeded(longToolName, false)
		assert.NoError(t, err, "Very long tool name should not cause error")
	})
}
