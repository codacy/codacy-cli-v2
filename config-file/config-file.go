package config_file

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ReadConfigFileWithLogging reads the configuration file and updates the global config with detailed logging
func ReadConfigFileWithLogging(configPath string) error {
	logger.Debug("Reading configuration file", logrus.Fields{
		"configPath": configPath,
	})

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Error("Configuration file does not exist", logrus.Fields{
			"configPath": configPath,
			"error":      err.Error(),
		})
		return err
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error("Failed to read configuration file", logrus.Fields{
			"configPath": configPath,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to read config file: %w", err)
	}

	logger.Debug("Parsing configuration file", logrus.Fields{
		"configPath": configPath,
		"fileSize":   len(data),
	})

	// Parse the YAML
	var configData struct {
		Runtimes []string `yaml:"runtimes"`
		Tools    []string `yaml:"tools"`
	}

	if err := yaml.Unmarshal(data, &configData); err != nil {
		logger.Error("Failed to parse configuration file", logrus.Fields{
			"configPath": configPath,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	logger.Debug("Configuration parsed successfully", logrus.Fields{
		"runtimes": configData.Runtimes,
		"tools":    configData.Tools,
	})

	// Convert the runtime strings to RuntimeConfig objects
	runtimeConfigs := make([]plugins.RuntimeConfig, 0, len(configData.Runtimes))
	for _, rt := range configData.Runtimes {
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
	toolConfigs := make([]plugins.ToolConfig, 0, len(configData.Tools))
	for _, tl := range configData.Tools {
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
