package tools

import (
	"codacy/cli-v2/domain"
	"codacy/cli-v2/plugins/tools/semgrep/embedded"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// semgrepRulesFile represents the structure of the rules.yaml file
type semgrepRulesFile struct {
	Rules []map[string]interface{} `yaml:"rules"`
}

// FilterRulesFromFile extracts enabled rules from a rules.yaml file based on configuration
func FilterRulesFromFile(rulesData []byte, config []domain.PatternConfiguration) ([]byte, error) {
	// Parse the YAML data
	var allRules semgrepRulesFile
	if err := yaml.Unmarshal(rulesData, &allRules); err != nil {
		return nil, fmt.Errorf("failed to parse rules file: %w", err)
	}

	// If no configuration provided, return all rules
	if len(config) == 0 {
		return rulesData, nil
	}

	// Create a map of enabled pattern IDs for faster lookup
	enabledPatterns := make(map[string]bool)
	for _, pattern := range config {
		if pattern.Enabled {
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

// GetSemgrepConfig gets the Semgrep configuration based on the pattern configuration.
// If no configuration is provided, returns all default rules.
func GetSemgrepConfig(config []domain.PatternConfiguration) ([]byte, error) {
	return FilterRulesFromFile(embedded.GetSemgrepRules(), config)
}

// GetDefaultSemgrepConfig gets the default Semgrep configuration
func GetDefaultSemgrepConfig() ([]byte, error) {
	// Return the embedded rules
	return embedded.GetSemgrepRules(), nil
}
