package tools

import (
	"os"
	"path/filepath"
)

// ConfigFileExists checks if a specific configuration file exists in the .codacy/tools-configs/ directory
// of the specified repository.
//
// Parameters:
//   - repositoryDir: The repository directory path
//   - fileName: The configuration file name to check for
//
// Returns:
//   - string: The relative path to the configuration file (for cmd args)
//   - bool: True if the file exists, false otherwise
func ConfigFileExists(repositoryDir string, fileName string) (string, bool) {
	configFile := filepath.Join(".codacy", "tools-configs", fileName)
	configFilePath := filepath.Join(repositoryDir, configFile)

	if _, err := os.Stat(configFilePath); err == nil {
		return configFile, true
	}

	return "", false
}
