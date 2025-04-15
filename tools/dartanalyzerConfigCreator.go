package tools

import (
	"codacy/cli-v2/domain"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

func CreateDartAnalyzerConfig(configuration []domain.PatternConfiguration) string {

	// Find Dart Analyzer patterns
	errorPatterns := []string{"ErrorProne", "Security", "Performance"}
	// Create analysis_options.yaml content
	config := map[string]interface{}{
		"analyzer": map[string]interface{}{
			"errors": map[string]string{},
		},
		"linter": map[string]interface{}{
			"rules": map[string]string{},
		},
	}

	errorsMap := config["analyzer"].(map[string]interface{})["errors"].(map[string]string)
	lintsMap := config["linter"].(map[string]interface{})["rules"].(map[string]string)
	for _, pattern := range configuration {
		if slices.Contains(errorPatterns, pattern.PatternDefinition.Category) {
			errorsMap[strings.TrimPrefix(pattern.PatternDefinition.Id, patternPrefix)] = strings.ToLower(pattern.PatternDefinition.Level)
		} else {
			lintsMap[strings.TrimPrefix(pattern.PatternDefinition.Id, patternPrefix)] = "true"
		}
	}

	// Write config to file
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return ""
	}
	return string(yamlData)
}
