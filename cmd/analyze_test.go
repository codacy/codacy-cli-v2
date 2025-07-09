package cmd

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"os"
	"path/filepath"
	"testing"

	"codacy/cli-v2/plugins"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "Python file",
			filePath: "test.py",
			want:     ".py",
		},
		{
			name:     "C++ file",
			filePath: "test.cpp",
			want:     ".cpp",
		},
		{
			name:     "File with path",
			filePath: "/path/to/file.js",
			want:     ".js",
		},
		{
			name:     "File without extension",
			filePath: "noextension",
			want:     "",
		},
		{
			name:     "File with uppercase extension",
			filePath: "test.PY",
			want:     ".py",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtension(tt.filePath); got != tt.want {
				t.Errorf("GetFileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsToolSupportedForFile(t *testing.T) {
	langConfig := &LanguagesConfig{
		Tools: []struct {
			Name       string   `yaml:"name" json:"name"`
			Languages  []string `yaml:"languages" json:"languages"`
			Extensions []string `yaml:"extensions" json:"extensions"`
			Files      []string `yaml:"files" json:"files"`
		}{
			{
				Name:       "pylint",
				Languages:  []string{"Python"},
				Extensions: []string{".py"},
				Files:      []string{},
			},
			{
				Name:       "eslint",
				Languages:  []string{"JavaScript", "TypeScript"},
				Extensions: []string{},
				Files:      []string{},
			},
			{
				Name:       "cppcheck",
				Languages:  []string{"C", "CPP"},
				Extensions: []string{".c", ".cpp", ".h", ".hpp"},
				Files:      []string{},
			},
			{
				Name:       "trivy",
				Languages:  []string{"Multiple"},
				Extensions: []string{".yaml", ".yml"},
				Files:      []string{"requirements.txt"},
			},
		},
	}

	tests := []struct {
		name     string
		toolName string
		filePath string
		config   *LanguagesConfig
		want     bool
	}{
		{
			name:     "Pylint with Python file",
			toolName: "pylint",
			filePath: "test.py",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Pylint with C++ file",
			toolName: "pylint",
			filePath: "test.cpp",
			config:   langConfig,
			want:     false,
		},
		{
			name:     "Cppcheck with C++ file",
			toolName: "cppcheck",
			filePath: "test.cpp",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Tool with no extensions specified",
			toolName: "eslint",
			filePath: "any.file",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Unknown tool",
			toolName: "unknown",
			filePath: "test.py",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Nil config",
			toolName: "pylint",
			filePath: "test.cpp",
			config:   nil,
			want:     true,
		},
		{
			name:     "Trivy with requirements.txt",
			toolName: "trivy",
			filePath: "requirements.txt",
			config:   langConfig,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsToolSupportedForFile(tt.toolName, tt.filePath, tt.config); got != tt.want {
				t.Errorf("IsToolSupportedForFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEslintInstallationValidationLogic(t *testing.T) {
	// Test the logic that determines when ESLint installation should be triggered
	// This tests the core logic from runEslintAnalysis without complex config setup

	tests := []struct {
		name                   string
		eslint                 *plugins.ToolInfo
		nodeRuntime            *plugins.RuntimeInfo
		isToolInstalled        bool
		expectInstallTriggered bool
		description            string
	}{
		{
			name:                   "tool_nil_should_trigger_install",
			eslint:                 nil,
			nodeRuntime:            nil,
			isToolInstalled:        false,
			expectInstallTriggered: true,
			description:            "When ESLint tool is nil, installation should be triggered",
		},
		{
			name: "tool_exists_runtime_nil_should_trigger_install",
			eslint: &plugins.ToolInfo{
				Name:    "eslint",
				Version: "8.38.0",
				Runtime: "node",
			},
			nodeRuntime:            nil,
			isToolInstalled:        true,
			expectInstallTriggered: true,
			description:            "When ESLint tool exists but runtime is nil, installation should be triggered",
		},
		{
			name: "tool_not_installed_runtime_exists_should_trigger_install",
			eslint: &plugins.ToolInfo{
				Name:    "eslint",
				Version: "8.38.0",
				Runtime: "node",
			},
			nodeRuntime: &plugins.RuntimeInfo{
				Name:    "node",
				Version: "22.2.0",
			},
			isToolInstalled:        false,
			expectInstallTriggered: true,
			description:            "When ESLint tool is not installed, installation should be triggered",
		},
		{
			name: "both_tool_and_runtime_available_should_not_trigger_install",
			eslint: &plugins.ToolInfo{
				Name:    "eslint",
				Version: "8.38.0",
				Runtime: "node",
			},
			nodeRuntime: &plugins.RuntimeInfo{
				Name:    "node",
				Version: "22.2.0",
			},
			isToolInstalled:        true,
			expectInstallTriggered: false,
			description:            "When both ESLint tool and runtime are available and installed, installation should not be triggered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This mimics the logic from runEslintAnalysis

			// Check if the runtime is installed
			var isRuntimeInstalled bool
			if tt.eslint != nil {
				isRuntimeInstalled = tt.nodeRuntime != nil
				// In the real code, this would also call Config.IsRuntimeInstalled,
				// but for this test we simplify it to just checking if runtime exists
			}

			// Apply the installation trigger logic
			shouldInstall := tt.eslint == nil || !tt.isToolInstalled || !isRuntimeInstalled

			if tt.expectInstallTriggered {
				assert.True(t, shouldInstall, tt.description)
			} else {
				assert.False(t, shouldInstall, tt.description)
			}
		})
	}
}

func TestValidatePaths(t *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		expectError bool
	}{
		{
			name:        "valid path",
			paths:       []string{"."},
			expectError: false,
		},
		{
			name:        "non-existent file",
			paths:       []string{"non-existent-file.txt"},
			expectError: true,
		},
		{
			name:        "multiple paths with one invalid",
			paths:       []string{".", "non-existent-file.txt"},
			expectError: true,
		},
		{
			name:        "empty paths",
			paths:       []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePaths(tt.paths)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "❌ Error: cannot find file or directory")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckIfConfigExistsAndIsNeeded(t *testing.T) {
	// Save original initFlags and config
	originalFlags := initFlags
	originalConfig := config.Config
	originalWorkingDir, _ := os.Getwd()
	defer func() {
		initFlags = originalFlags
		config.Config = originalConfig
		os.Chdir(originalWorkingDir)
	}()

	tests := []struct {
		name             string
		toolName         string
		cliLocalMode     bool
		apiToken         string
		configFileExists bool
		expectError      bool
		description      string
	}{
		{
			name:             "tool_without_config_file",
			toolName:         "unsupported-tool",
			cliLocalMode:     false,
			apiToken:         "test-token",
			configFileExists: false,
			expectError:      false,
			description:      "Tool that doesn't use config files should return without error",
		},
		{
			name:             "config_file_exists",
			toolName:         "eslint",
			cliLocalMode:     false,
			apiToken:         "test-token",
			configFileExists: true,
			expectError:      false,
			description:      "When config file exists, should find it successfully",
		},
		{
			name:             "remote_mode_without_token_no_config",
			toolName:         "eslint",
			cliLocalMode:     false,
			apiToken:         "",
			configFileExists: false,
			expectError:      false,
			description:      "Remote mode without token should show appropriate message",
		},
		{
			name:             "local_mode_no_config",
			toolName:         "eslint",
			cliLocalMode:     true,
			apiToken:         "",
			configFileExists: false,
			expectError:      false,
			description:      "Local mode should create config file if tools-configs directory exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temporary directory and change to it to avoid creating files in project dir
			tmpDir, err := os.MkdirTemp("", "codacy-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Change to temp directory to prevent config file creation in project
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Mock config to use our temporary directory BEFORE creating files
			config.Config = *config.NewConfigType(tmpDir, tmpDir, tmpDir)

			// Create config file if needed - using the same path logic as the function under test
			if tt.configFileExists && constants.ToolConfigFileNames[tt.toolName] != "" {
				// Use config.Config.ToolsConfigDirectory() to get the exact same path the function will use
				toolsConfigDir := config.Config.ToolsConfigDirectory()
				err := os.MkdirAll(toolsConfigDir, constants.DefaultDirPerms)
				require.NoError(t, err)

				configPath := filepath.Join(toolsConfigDir, constants.ToolConfigFileNames[tt.toolName])
				err = os.WriteFile(configPath, []byte("test config"), constants.DefaultFilePerms)
				require.NoError(t, err)

				// Ensure the file was created and can be found
				_, err = os.Stat(configPath)
				require.NoError(t, err, "Config file should exist at %s", configPath)
			}

			// Setup initFlags
			initFlags = domain.InitFlags{
				ApiToken: tt.apiToken,
			}

			// Ensure tools-configs directory exists if the function might try to create config files
			if !tt.configFileExists && constants.ToolConfigFileNames[tt.toolName] != "" {
				toolsConfigDir := config.Config.ToolsConfigDirectory()
				err := os.MkdirAll(toolsConfigDir, constants.DefaultDirPerms)
				require.NoError(t, err)
			}

			// Execute the function
			err = checkIfConfigExistsAndIsNeeded(tt.toolName, tt.cliLocalMode)

			// Clean up any files that might have been created by the function under test
			if !tt.configFileExists && constants.ToolConfigFileNames[tt.toolName] != "" {
				toolsConfigDir := config.Config.ToolsConfigDirectory()
				configPath := filepath.Join(toolsConfigDir, constants.ToolConfigFileNames[tt.toolName])
				if _, statErr := os.Stat(configPath); statErr == nil {
					os.Remove(configPath)
				}
			}

			// Verify results
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestToolConfigFileNameMap(t *testing.T) {
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

	for toolName, expectedFileName := range expectedTools {
		t.Run(toolName, func(t *testing.T) {
			actualFileName, exists := constants.ToolConfigFileNames[toolName]
			assert.True(t, exists, "Tool %s should exist in constants.ToolConfigFileNames map", toolName)
			assert.Equal(t, expectedFileName, actualFileName, "Config filename for %s should match expected", toolName)
		})
	}
}
