package config_file

import (
	"codacy/cli-v2/config"
	"gopkg.in/yaml.v3"
	"os"
)

type configFile struct {
	RUNTIMES []string
	TOOLS    []string
}

func parseConfigFile(configContents []byte) error {
	configFile := configFile{}
	if err := yaml.Unmarshal(configContents, &configFile); err != nil {
		return err
	}

	for _, rt := range configFile.RUNTIMES {
		ct, err := parseConfigTool(rt)
		if err != nil {
			return err
		}
		config.Config.AddRuntime(config.NewRuntime(ct.name, ct.version))
	}

	for _, tl := range configFile.TOOLS {
		ct, err := parseConfigTool(tl)
		if err != nil {
			return err
		}
		config.Config.AddTool(config.NewRuntime(ct.name, ct.version))
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