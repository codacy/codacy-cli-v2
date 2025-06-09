package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codacy/cli-v2/config"
	"codacy/cli-v2/utils/logger"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// validateCodacyYAML validates that codacy.yaml exists and is not empty
// Returns an error if validation fails, nil if validation passes
func validateCodacyYAML() error {
	codacyYAMLPath := config.Config.ProjectConfigFile()

	// Check if file exists
	if _, err := os.Stat(codacyYAMLPath); os.IsNotExist(err) {
		logger.Error("codacy.yaml file not found", logrus.Fields{
			"file": codacyYAMLPath,
		})
		return fmt.Errorf("❌ Fatal: codacy.yaml file not found. Please run 'codacy-cli init' to initialize your configuration first")
	} else if err != nil {
		logger.Error("Error accessing codacy.yaml file", logrus.Fields{
			"file":  codacyYAMLPath,
			"error": err,
		})
		return fmt.Errorf("error accessing %s: %w", codacyYAMLPath, err)
	}

	// Read file content
	content, err := os.ReadFile(codacyYAMLPath)
	if err != nil {
		logger.Error("Failed to read codacy.yaml file", logrus.Fields{
			"file":  codacyYAMLPath,
			"error": err,
		})
		return fmt.Errorf("error reading %s: %w", codacyYAMLPath, err)
	}

	// Check if file is empty or contains only whitespace
	if len(strings.TrimSpace(string(content))) == 0 {
		logger.Error("codacy.yaml file is empty", logrus.Fields{
			"file": codacyYAMLPath,
		})
		return fmt.Errorf("❌ Fatal: %s is empty. Please run 'codacy-cli init' to initialize your configuration first", filepath.Base(codacyYAMLPath))
	}

	// Try to parse YAML to ensure it's valid
	var configData map[string]interface{}
	if err := yaml.Unmarshal(content, &configData); err != nil {
		logger.Error("Failed to parse codacy.yaml file", logrus.Fields{
			"file":  codacyYAMLPath,
			"error": err,
		})
		if strings.Contains(err.Error(), "cannot unmarshal") {
			return fmt.Errorf("❌ Fatal: %s contains invalid configuration - run 'codacy-cli config reset' to fix: %v", filepath.Base(codacyYAMLPath), err)
		}
		return fmt.Errorf("❌ Fatal: %s is broken or has invalid YAML format - run 'codacy-cli config reset' to reinitialize your configuration", filepath.Base(codacyYAMLPath))
	}

	// Ensure configData is not nil after unmarshaling
	if configData == nil {
		logger.Error("codacy.yaml unmarshaled to nil data", logrus.Fields{
			"file": codacyYAMLPath,
		})
		return fmt.Errorf("❌ Fatal: %s contains no valid configuration. Please run 'codacy-cli init' to initialize your configuration first", filepath.Base(codacyYAMLPath))
	}

	return nil
}

// shouldSkipValidation checks if the current command should skip codacy.yaml validation
func shouldSkipValidation(cmdName string) bool {
	skipCommands := []string{
		"init",
		"help",
		"version",
		"reset",      // config reset should work even with empty/invalid codacy.yaml
		"codacy-cli", // root command when called without subcommands
	}

	for _, skipCmd := range skipCommands {
		if cmdName == skipCmd {
			return true
		}
	}

	return false
}
