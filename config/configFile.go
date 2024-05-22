package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type configFile struct {
	RUNTIMES []string
	TOOLS    []string
}

func parseConfigFile(configContents []byte) (map[string]*Runtime, error) {
	configFile := configFile{}
	if err := yaml.Unmarshal(configContents, &configFile); err != nil {
		return nil, err
	}

	runtimes := make(map[string]*Runtime)
	for _, rt := range configFile.RUNTIMES {
		ct, err := parseConfigTool(rt)
		if err != nil {
			return nil, err
		}
		runtimes[ct.name] = &Runtime{
			name: ct.name,
			version: ct.version,
		}
	}

	for _, tl := range configFile.TOOLS {
		ct, err := parseConfigTool(tl)
		if err != nil {
			return nil, err
		}
		switch ct.name {
		case "eslint":
			runtimes["node"].AddTool(ct)
		}
	}

	return runtimes, nil
}

func ReadConfigFile(configPath string) (map[string]*Runtime, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	return parseConfigFile(content)
}