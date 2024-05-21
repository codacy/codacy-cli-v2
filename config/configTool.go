package config

import (
	"errors"
	"strings"
)

type ConfigTool struct {
	name string
	version string
}

func (ct *ConfigTool) Name() string {
	return ct.name
}

func (ct *ConfigTool) Version() string {
	return ct.version
}

func parseConfigTool(tool string) (*ConfigTool, error) {
	toolSplited := strings.Split(tool, "@")
	if len(toolSplited) != 2 {
		return &ConfigTool{}, errors.New("invalid tool format")
	}

	return &ConfigTool{name: toolSplited[0], version: toolSplited[1]}, nil
}