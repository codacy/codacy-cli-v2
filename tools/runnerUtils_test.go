package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFileExists(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()

	// Create .codacy/tools-configs directory
	configDir := filepath.Join(tempDir, ".codacy", "tools-configs")

	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err, "Failed to create test directory structure")

	// Create a test config file
	testConfigFile := filepath.Join(configDir, "test-config.yaml")
	err = os.WriteFile(testConfigFile, []byte("test content"), 0644)
	assert.NoError(t, err, "Failed to create test config file")

	// Test case 1: Config file exists
	configPath, exists := ConfigFileExists(tempDir, "test-config.yaml")
	assert.True(t, exists, "Config file should exist")
	assert.Equal(t, filepath.Join(".codacy", "tools-configs", "test-config.yaml"), configPath,
		"Config path should be correctly formed relative path")

	// Test case 2: Config file doesn't exist
	configPath, exists = ConfigFileExists(tempDir, "non-existent-config.yaml")
	assert.False(t, exists, "Config file should not exist")
	assert.Equal(t, "", configPath, "Config path should be empty for non-existent file")
}
