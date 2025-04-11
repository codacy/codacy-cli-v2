package tools

import (
	"codacy/cli-v2/config"
	"os"
	"path/filepath"
)

// ConfigFileExists checks if any of the specified configuration files exist in the .codacy/tools-configs/
// or on the root of the repository directory.
//
// Parameters:
//   - conf: The configuration object containing the tools config directory
//   - fileNames: A list of configuration file names to check for
//
// Returns:
//   - string: The relative path to the first configuration file found (for cmd args)
//   - bool: True if any file exists, false otherwise
func ConfigFileExists(conf config.ConfigType, fileNames ...string) (string, bool) {
	for _, fileName := range fileNames {
		generatedConfigFile := filepath.Join(conf.ToolsConfigDirectory(), fileName)
		existingConfigFile := filepath.Join(conf.RepositoryDirectory(), fileName)

		if _, err := os.Stat(generatedConfigFile); err == nil {
			return generatedConfigFile, true
		} else if _, err := os.Stat(existingConfigFile); err == nil {
			return existingConfigFile, true
		}
	}

	return "", false
}
