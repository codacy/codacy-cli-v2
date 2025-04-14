package tools

import (
	"codacy/cli-v2/domain"
	"fmt"
	"slices"
	"strconv"
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
		fmt.Println(pattern.PatternDefinition.Id, pattern.Enabled, pattern.PatternDefinition.Category)
		if slices.Contains(errorPatterns, pattern.PatternDefinition.Category) {
			if pattern.Enabled {
				errorsMap[strings.TrimPrefix(pattern.PatternDefinition.Id, patternPrefix)] = strings.ToLower(pattern.PatternDefinition.Level)
			} else {
				errorsMap[strings.TrimPrefix(pattern.PatternDefinition.Id, patternPrefix)] = "ignore"
			}

		} else {
			lintsMap[strings.TrimPrefix(pattern.PatternDefinition.Id, patternPrefix)] = strconv.FormatBool(pattern.Enabled)
		}
	}

	// Write config to file
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return ""
	}
	return string(yamlData)
}
