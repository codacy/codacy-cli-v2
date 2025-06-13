package cmd

import (
	"codacy/cli-v2/plugins"
	"testing"

	"github.com/stretchr/testify/assert"
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
		}{
			{
				Name:       "pylint",
				Languages:  []string{"Python"},
				Extensions: []string{".py"},
			},
			{
				Name:       "cppcheck",
				Languages:  []string{"C", "CPP"},
				Extensions: []string{".c", ".cpp", ".h", ".hpp"},
			},
			{
				Name:       "trivy",
				Languages:  []string{"Multiple"},
				Extensions: []string{},
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
			toolName: "trivy",
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
