package config_file

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/plugins"
	"os"

	"gopkg.in/yaml.v3"
)

type configFile struct {
	RUNTIMES []string `yaml:"runtimes"`
	TOOLS    []string `yaml:"tools"`
}

func parseConfigFile(configContents []byte) error {
	configFile := configFile{}
	if err := yaml.Unmarshal(configContents, &configFile); err != nil {
		return err
	}

	// Convert the runtime strings to RuntimeConfig objects
	runtimeConfigs := make([]plugins.RuntimeConfig, 0, len(configFile.RUNTIMES))
	for _, rt := range configFile.RUNTIMES {
		ct, err := parseConfigTool(rt)
		if err != nil {
			return err
		}
		runtimeConfigs = append(runtimeConfigs, plugins.RuntimeConfig{
			Name:    ct.name,
			Version: ct.version,
		})
	}

	// Add all runtimes at once
	if err := config.Config.AddRuntimes(runtimeConfigs); err != nil {
		return err
	}

	// Convert the tool strings to ToolConfig objects
	toolConfigs := make([]plugins.ToolConfig, 0, len(configFile.TOOLS))
	for _, tl := range configFile.TOOLS {
		ct, err := parseConfigTool(tl)
		if err != nil {
			return err
		}
		toolConfigs = append(toolConfigs, plugins.ToolConfig{
			Name:    ct.name,
			Version: ct.version,
		})
	}

	// Add all tools at once
	if err := config.Config.AddTools(toolConfigs); err != nil {
		return err
	}

	return nil
}

func ReadConfigFile(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return parseConfigFile(content)
}
