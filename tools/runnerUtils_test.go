package tools

import (
	"codacy/cli-v2/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFileExistsInToolsConfigDirectory(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "src")
	repositoryCache := filepath.Join(repoDir, ".codacy")

	// Create configuration
	config := *config.NewConfigType(repoDir, repositoryCache, "unused-global-cache")

	// Create .codacy/tools-configs directory
	configDir := filepath.Join(repoDir, ".codacy", "tools-configs")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err, "Failed to create test directory structure")

	// Create a test config file on the configDir
	generatedConfigFile := filepath.Join(configDir, "generated-config.yaml")
	err = os.WriteFile(generatedConfigFile, []byte("test content"), 0644)
	assert.NoError(t, err, "Failed to create test config file")

	// Test case: Config file exists in tools config directory
	configPath, exists := ConfigFileExists(config, "generated-config.yaml")
	assert.True(t, exists, "Config file should exist in tools config directory")
	assert.Equal(t, filepath.Join(config.ToolsConfigDirectory(), "generated-config.yaml"), configPath,
		"Config path should be correctly formed relative path")
}

func TestConfigFileExistsInRepositoryDirectory(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "src")
	repositoryCache := filepath.Join(repoDir, ".codacy")

	// Create configuration
	config := *config.NewConfigType(repoDir, repositoryCache, "unused-global-cache")

	// Create .codacy/tools-configs directory
	configDir := filepath.Join(repoDir, ".codacy", "tools-configs")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err, "Failed to create test directory structure")

	// Create a test config file on the repository directory
	existingConfigFile := filepath.Join(repoDir, "existing-config.yaml")
	err = os.WriteFile(existingConfigFile, []byte("test content"), 0644)
	assert.NoError(t, err, "Failed to create test config file")

	// Test case: The existing config file gets picked up
	configPath, exists := ConfigFileExists(config, "existing-config.yaml")
	assert.True(t, exists, "Config file should exist in tools config directory")
	assert.Equal(t, filepath.Join(config.RepositoryDirectory(), "existing-config.yaml"), configPath,
		"Config path should be correctly formed relative path")
}

func TestConfigFileDoesNotExist(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "src")
	repositoryCache := filepath.Join(repoDir, ".codacy")

	// Create configuration
	config := *config.NewConfigType(repoDir, repositoryCache, "unused-global-cache")

	// Test case: Config file does not exist
	configPath, exists := ConfigFileExists(config, "non-existent-config.yaml")
	assert.False(t, exists, "Config file should not exist")
	assert.Equal(t, "", configPath, "Config path should be empty for non-existent file")
}
