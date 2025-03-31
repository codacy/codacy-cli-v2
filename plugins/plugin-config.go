package plugins

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PluginConfig represents the structure of plugin.yaml files
type PluginConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Downloads   []DownloadConfig  `yaml:"downloads"`
	ArchMapping map[string]string `yaml:"arch_mapping"`
	Binaries    []BinaryConfig    `yaml:"binaries"`
}

// DownloadConfig represents the download configuration in plugin.yaml
type DownloadConfig struct {
	OS               []string        `yaml:"os"`
	URLTemplate      string          `yaml:"url_template"`
	FileNameTemplate string          `yaml:"file_name_template"`
	Extension        ExtensionConfig `yaml:"extension"`
}

// ExtensionConfig defines the file extension based on OS
type ExtensionConfig struct {
	Windows string `yaml:"windows"`
	Default string `yaml:"default"`
}

// BinaryConfig represents a binary executable
type BinaryConfig struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// LoadPluginConfig loads a plugin configuration from the given path
func LoadPluginConfig(path string) (*PluginConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading plugin config file: %w", err)
	}

	var config PluginConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing plugin config file: %w", err)
	}

	return &config, nil
}
