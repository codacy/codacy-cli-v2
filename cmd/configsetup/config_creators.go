package configsetup

import (
	"fmt"
	"path/filepath"

	"codacy/cli-v2/config"
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/tools"

	"gopkg.in/yaml.v3"
)

func CreateLanguagesConfigFileLocal(toolsConfigDir string) error {
	// Build tool language configurations from API
	configTools, err := tools.BuildLanguagesConfigFromAPI()
	if err != nil {
		return fmt.Errorf("failed to build languages config from API: %w", err)
	}

	// Create the config structure
	config := domain.LanguagesConfig{
		Tools: configTools,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal languages config to YAML: %w", err)
	}

	return writeConfigFile(filepath.Join(toolsConfigDir, constants.LanguagesConfigFileName), data)
}

func CreateGitIgnoreFile() error {
	gitIgnorePath := filepath.Join(config.Config.LocalCodacyDirectory(), constants.GitIgnoreFileName)
	content := "# Codacy CLI\ntools-configs/\n.gitignore\ncli-config.yaml\nlogs/\n"
	return writeConfigFile(gitIgnorePath, []byte(content))
}

func CreateConfigurationFiles(tools []domain.Tool, cliLocalMode bool, flags domain.InitFlags) error {
	// Create project config file
	configContent := ConfigFileTemplate(tools)
	if err := writeConfigFile(config.Config.ProjectConfigFile(), []byte(configContent)); err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	// Create CLI config file
	cliConfigContent := buildCliConfigContent(cliLocalMode, flags)
	if err := writeConfigFile(config.Config.CliConfigFile(), []byte(cliConfigContent)); err != nil {
		return fmt.Errorf("failed to write CLI config file: %w", err)
	}

	return nil
}

// buildCliConfigContent creates the CLI configuration content
func buildCliConfigContent(cliLocalMode bool, initFlags domain.InitFlags) string {
	if cliLocalMode {
		return fmt.Sprintf("mode: local")
	} else {
		return fmt.Sprintf("mode: remote\nprovider: %s\norganization: %s\nrepository: %s", initFlags.Provider, initFlags.Organization, initFlags.Repository)
	}
}
