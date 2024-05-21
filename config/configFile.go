package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type configFile struct {
	RUNTIMES []string
	TOOLS    []string
}

type Runtime struct {
	name string
	version string
	tools []ConfigTool
}

func (r *Runtime) Name() string {
	return r.name
}

func (r *Runtime) Version() string {
	return r.version
}

func (r *Runtime) Tools() []ConfigTool {
	return r.tools
}

func (r *Runtime) AddTool(tool *ConfigTool) {
	r.tools = append(r.tools, *tool)
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
		}
		runtimes[ct.name] = &Runtime{
			name: ct.name,
			version: ct.version,
		}
	}

	for _, tl := range configFile.TOOLS {
		ct, err := parseConfigTool(tl)
		if err != nil {
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