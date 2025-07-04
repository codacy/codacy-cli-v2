package lizard

import (
	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// CreateLizardConfig generates a Lizard configuration file content based on the API configuration
func CreateLizardConfig(toolsConfigDir string, patterns []domain.PatternDefinition) error {
	patternInfo := make(map[string]map[string]interface{})

	for _, pattern := range patterns {
		metricType := getMetricTypeFromPatternId(pattern.Id)
		if metricType == "" {
			fmt.Printf("Warning: Invalid pattern ID format: %s\n", pattern.Id)
			continue
		}

		threshold := getThresholdFromParams(pattern.Parameters)
		if threshold != 0 {
			// Create a unique key for this pattern that includes the severity
			patternKey := pattern.Id
			patternInfo[patternKey] = map[string]interface{}{
				"id":            pattern.Id,
				"category":      pattern.Category,
				"level":         pattern.Level,
				"severityLevel": pattern.SeverityLevel,
				"title":         pattern.Title,
				"description":   pattern.Description,
				"explanation":   pattern.Explanation,
				"timeToFix":     pattern.TimeToFix,
				"threshold":     threshold,
			}
		}
	}

	config := map[string]interface{}{
		"patterns": patternInfo,
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return os.WriteFile(filepath.Join(toolsConfigDir, "lizard.yaml"), yamlData, constants.DefaultFilePerms)
}

// getThresholdFromParams extracts the threshold value from the parameters
func getThresholdFromParams(params []domain.ParameterConfiguration) int {
	for _, param := range params {
		if param.Name == "threshold" {
			if param.Value != "" {
				threshold, err := strconv.Atoi(param.Value)
				if err == nil {
					return threshold
				}
			} else if param.Default != "" {
				threshold, err := strconv.Atoi(param.Default)
				if err == nil {
					return threshold
				}
			}
		}
	}
	return 0
}

// getMetricTypeFromPatternId extracts the metric type from the pattern ID
func getMetricTypeFromPatternId(patternID string) string {
	// Pattern IDs are in the format "Lizard_metric-severity"

	parts := strings.Split(patternID, "_")
	if len(parts) != 2 {
		fmt.Printf("Warning: Invalid pattern ID format: %s\n", patternID)
		return ""
	}

	// Extract the metric parts from the second part
	metricParts := strings.Split(parts[1], "-")
	if len(metricParts) < 2 {
		fmt.Printf("Warning: Invalid metric format: %s\n", parts[1])
		return ""
	}

	// The last part is always the severity (medium, critical, etc.)
	// Everything before that is the metric type
	metricType := strings.Join(metricParts[:len(metricParts)-1], "-")

	// Validating that the metric type is one of the known types
	switch metricType {
	case "ccn":
		return "ccn"
	case "nloc":
		return "nloc"
	case "file-nloc":
		return "file-nloc"
	case "parameter-count":
		return "parameter-count"
	default:
		fmt.Printf("Warning: Unknown metric type: %s\n", metricType)
		return ""
	}
}

// ReadConfig reads and parses the Lizard configuration file
func ReadConfig(configPath string) ([]domain.PatternDefinition, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	var patterns []domain.PatternDefinition
	for _, pattern := range config["patterns"] {
		threshold, ok := pattern["threshold"].(int)
		if !ok {
			continue
		}

		patterns = append(patterns, domain.PatternDefinition{
			Id:            pattern["id"].(string),
			Category:      pattern["category"].(string),
			Level:         pattern["level"].(string),
			SeverityLevel: pattern["severityLevel"].(string),
			Title:         pattern["title"].(string),
			Description:   pattern["description"].(string),
			Explanation:   pattern["explanation"].(string),
			TimeToFix:     pattern["timeToFix"].(int),
			Parameters: []domain.ParameterConfiguration{
				{
					Name:    "threshold",
					Value:   fmt.Sprintf("%d", threshold),
					Default: fmt.Sprintf("%d", threshold),
				},
			},
		})
	}

	return patterns, nil
}
