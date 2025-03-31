package config

import (
	"codacy/cli-v2/plugins"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddTools(t *testing.T) {
	// Set up a temporary config for testing
	originalConfig := Config
	defer func() { Config = originalConfig }() // Restore original config after test
	
	tempDir, err := os.MkdirTemp("", "codacy-tools-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Initialize config with test directories
	Config = ConfigType{
		toolsDirectory: tempDir,
		tools:          make(map[string]*plugins.ToolInfo),
	}
	
	// Create a list of tool configs for testing
	configs := []plugins.ToolConfig{
		{
			Name:    "eslint",
			Version: "8.38.0",
		},
	}
	
	// Add tools to the config
	err = Config.AddTools(configs)
	assert.NoError(t, err)
	
	// Assert we have the expected tool in the config
	assert.Contains(t, Config.Tools(), "eslint")
	
	// Get the eslint tool info
	eslintInfo := Config.Tools()["eslint"]
	
	// Assert the basic tool info is correct
	assert.Equal(t, "eslint", eslintInfo.Name)
	assert.Equal(t, "8.38.0", eslintInfo.Version)
	assert.Equal(t, "node", eslintInfo.Runtime)
	
	// Assert the install directory is correct
	expectedInstallDir := filepath.Join(tempDir, "eslint@8.38.0")
	assert.Equal(t, expectedInstallDir, eslintInfo.InstallDir)
}

func TestExecuteToolTemplate(t *testing.T) {
	// Test template execution with different data
	templateStr := "install --prefix {{.InstallDir}} {{.PackageName}}@{{.Version}}"
	data := map[string]string{
		"InstallDir":  "/test/tools/eslint@8.38.0",
		"PackageName": "eslint",
		"Version":     "8.38.0",
	}

	result, err := executeToolTemplate(templateStr, data)
	assert.NoError(t, err)
	assert.Equal(t, "install --prefix /test/tools/eslint@8.38.0 eslint@8.38.0", result)

	// Test conditional registry template
	registryTemplateStr := "{{if .Registry}}config set registry {{.Registry}}{{end}}"
	
	// With registry
	dataWithRegistry := map[string]string{
		"Registry": "https://registry.npmjs.org/",
	}
	resultWithRegistry, err := executeToolTemplate(registryTemplateStr, dataWithRegistry)
	assert.NoError(t, err)
	assert.Equal(t, "config set registry https://registry.npmjs.org/", resultWithRegistry)
	
	// Without registry
	dataWithoutRegistry := map[string]string{}
	resultWithoutRegistry, err := executeToolTemplate(registryTemplateStr, dataWithoutRegistry)
	assert.NoError(t, err)
	assert.Equal(t, "", resultWithoutRegistry)
}