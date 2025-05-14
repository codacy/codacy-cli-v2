package tools

import (
	"codacy/cli-v2/domain"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// semgrepRulesFile represents the structure of the rules.yaml file
type semgrepRulesFile struct {
	Rules []map[string]interface{} `yaml:"rules"`
}

// getExecutablePath is a variable that holds the function to get the executable path
// This is used for testing purposes
var getExecutablePath = os.Executable

// FilterRulesFromFile extracts enabled rules from a rules.yaml file based on configuration
func FilterRulesFromFile(rulesFilePath string, config []domain.PatternConfiguration) ([]byte, error) {
	// Read the rules.yaml file
	data, err := os.ReadFile(rulesFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	// Parse the YAML file
	var allRules semgrepRulesFile
	if err := yaml.Unmarshal(data, &allRules); err != nil {
		return nil, fmt.Errorf("failed to parse rules file: %w", err)
	}

	// Create a map of enabled pattern IDs for faster lookup
	enabledPatterns := make(map[string]bool)
	for _, pattern := range config {
		if pattern.Enabled && pattern.PatternDefinition.Enabled {
			// Extract rule ID from pattern ID
			parts := strings.SplitN(pattern.PatternDefinition.Id, "_", 2)
			if len(parts) == 2 {
				ruleID := parts[1]
				enabledPatterns[ruleID] = true
			}
		}
	}

	// Filter the rules based on enabled patterns
	var filteredRules semgrepRulesFile
	filteredRules.Rules = []map[string]interface{}{}

	for _, rule := range allRules.Rules {
		// Get the rule ID
		if ruleID, ok := rule["id"].(string); ok && enabledPatterns[ruleID] {
			// If this rule is enabled, include it
			filteredRules.Rules = append(filteredRules.Rules, rule)
		}
	}

	// If no rules match, return an error
	if len(filteredRules.Rules) == 0 {
		return nil, fmt.Errorf("no matching rules found")
	}

	// Marshal the filtered rules back to YAML
	return yaml.Marshal(filteredRules)
}

// GetSemgrepConfig gets the Semgrep configuration based on the pattern configuration
func GetSemgrepConfig(config []domain.PatternConfiguration) ([]byte, error) {
	// Get the executable's directory
	execPath, err := getExecutablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Get the default rules file location relative to the executable
	rulesFile := filepath.Join(execDir, "plugins", "tools", "semgrep", "rules.yaml")

	// Check if it exists and config is not empty
	if _, err := os.Stat(rulesFile); err == nil && len(config) > 0 {
		// Try to filter rules from the file
		return FilterRulesFromFile(rulesFile, config)
	}

	// If rules.yaml doesn't exist or config is empty, return an error
	return nil, fmt.Errorf("rules.yaml not found or empty configuration")
}

// GetDefaultSemgrepConfig gets the default Semgrep configuration
func GetDefaultSemgrepConfig() ([]byte, error) {
	// Get the executable's directory
	execPath, err := getExecutablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Get the default rules file location relative to the executable
	rulesFile := filepath.Join(execDir, "plugins", "tools", "semgrep", "rules.yaml")

	// If the file exists, return its contents
	if _, err := os.Stat(rulesFile); err == nil {
		return os.ReadFile(rulesFile)
	}

	// Return an error if rules.yaml doesn't exist
	return nil, fmt.Errorf("rules.yaml not found")
}
