package tools

import (
	"codacy/cli-v2/config"
	"os"
	"path/filepath"
)

// ConfigFileExists checks if a specific configuration file exists in the .codacy/tools-configs/
// or on the root of the repository directory.
//
// Parameters:
//   - conf: The configuration object containing the tools config directory
//   - fileName: The configuration file name to check for
//
// Returns:
//   - string: The relative path to the configuration file (for cmd args)
//   - bool: True if the file exists, false otherwise
func ConfigFileExists(conf config.ConfigType, fileName string) (string, bool) {
	generatedConfigFile := filepath.Join(conf.ToolsConfigDirectory(), fileName)
	existingConfigFile := filepath.Join(conf.RepositoryDirectory(), fileName)

	if _, err := os.Stat(generatedConfigFile); err == nil {
		return generatedConfigFile, true
	} else if _, err := os.Stat(existingConfigFile); err == nil {
		return existingConfigFile, true
	}

	return "", false
}
