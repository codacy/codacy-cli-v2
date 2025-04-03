package tools

import (
	"codacy/cli-v2/tools/types"
	"encoding/json"
	"fmt"
	"strings"
)

// ParsePylintPatternsFromJSON parses a JSON array of Pylint patterns into types.PylintPatternConfiguration array
func ParsePylintPatternsFromJSON(jsonData []byte) ([]types.PylintPatternConfiguration, error) {
	var response struct {
		Data []struct {
			PatternDefinition struct {
				ID         string `json:"id"`
				Parameters []struct {
					Name    string `json:"name"`
					Default string `json:"default"`
				} `json:"parameters"`
			} `json:"patternDefinition"`
			Enabled    bool `json:"enabled"`
			Parameters []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"parameters"`
		} `json:"data"`
	}

	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, err
	}

	patterns := make([]types.PylintPatternConfiguration, len(response.Data))
	for i, pattern := range response.Data {
		defaultParameters := make([]types.PylintPatternParameterConfiguration, len(pattern.PatternDefinition.Parameters))
		for j, param := range pattern.PatternDefinition.Parameters {
			defaultParameters[j] = types.PylintPatternParameterConfiguration{
				Name:        param.Name,
				Value:       param.Default,
				SectionName: GetParameterSection(param.Name),
			}
		}

		parameters := make([]types.PylintPatternParameterConfiguration, len(pattern.Parameters))
		for j, param := range pattern.Parameters {
			parameters[j] = types.PylintPatternParameterConfiguration{
				Name:        param.Name,
				Value:       param.Value,
				SectionName: GetParameterSection(param.Name),
			}
		}

		patterns[i] = types.PylintPatternConfiguration{
			Id:         ExtractPatternID(pattern.PatternDefinition.ID),
			Enabled:    pattern.Enabled,
			Parameters: setConfigurationParameters(parameters, defaultParameters),
		}
	}

	return patterns, nil
}

// FilterEnabledPatterns returns only the enabled patterns from the given slice
func FilterEnabledPatterns(patterns []types.PylintPatternConfiguration) []types.PylintPatternConfiguration {
	enabledPatterns := make([]types.PylintPatternConfiguration, 0)
	for _, pattern := range patterns {
		if pattern.Enabled {
			enabledPatterns = append(enabledPatterns, pattern)
		}
	}
	return enabledPatterns
}

// DefaultPylintParameters contains the default values for common Pylint parameters
var DefaultPylintParameters = map[string]string{
	"max-line-length":    "100",
	"max-doc-length":     "100",
	"max-args":           "5",
	"max-attributes":     "7",
	"max-branches":       "12",
	"max-locals":         "15",
	"max-parents":        "7",
	"max-public-methods": "20",
	"max-returns":        "6",
	"max-statements":     "50",
}

// ExtractPatternID returns the part of the pattern ID after the underscore
// For example: "PyLintPython3_C0301" -> "C0301"
func ExtractPatternID(fullID string) string {
	parts := strings.Split(fullID, "_")
	if len(parts) > 1 {
		return parts[1]
	}
	return fullID
}

// GeneratePylintRC generates a pylintrc file content with the specified patterns enabled
func GeneratePylintRC(patterns []types.PylintPatternConfiguration) string {
	var rcContent strings.Builder

	// Write header
	rcContent.WriteString("[MASTER]\n")
	rcContent.WriteString("ignore=CVS\n")
	rcContent.WriteString("persistent=yes\n")
	rcContent.WriteString("load-plugins=\n\n")

	// Disable all patterns by default
	rcContent.WriteString("[MESSAGES CONTROL]\n")
	rcContent.WriteString("disable=all\n")

	// Collect all enabled pattern IDs
	var enabledPatterns []string
	for _, pattern := range patterns {
		if pattern.Enabled {
			enabledPatterns = append(enabledPatterns, ExtractPatternID(pattern.Id))
		}
	}

	// Write all enabled patterns in a single line
	if len(enabledPatterns) > 0 {
		rcContent.WriteString(fmt.Sprintf("enable=%s\n", strings.Join(enabledPatterns, ",")))
	}
	rcContent.WriteString("\n")

	// Group parameters by section
	groupedParams := GroupParametersBySection(patterns)

	// Write parameters for each section
	for sectionName, params := range groupedParams {
		rcContent.WriteString(fmt.Sprintf("[%s]\n", sectionName))
		for _, param := range params {
			value := param.Value
			if value == "" {
				// If no value is set, use default from DefaultPylintParameters
				if defaultVal, ok := DefaultPylintParameters[param.Name]; ok {
					value = defaultVal
				}
			}
			rcContent.WriteString(fmt.Sprintf("%s=%s\n", param.Name, value))
		}
		rcContent.WriteString("\n")
	}

	return rcContent.String()
}

// setConfigurationParameters returns the first array if it's not empty, otherwise returns the second array
func setConfigurationParameters(parameters, defaultParameters []types.PylintPatternParameterConfiguration) []types.PylintPatternParameterConfiguration {
	if len(parameters) > 0 {
		return parameters
	}
	return defaultParameters
}

// groupParametersBySection groups parameters by their section name
func GroupParametersBySection(patterns []types.PylintPatternConfiguration) map[string][]types.PylintPatternParameterConfiguration {
	// Initialize the result map
	groupedParams := make(map[string][]types.PylintPatternParameterConfiguration)

	// Iterate through each pattern
	for _, pattern := range patterns {
		// Iterate through each parameter
		for _, param := range pattern.Parameters {
			// Skip parameters without a section name
			if param.SectionName == nil {
				continue
			}

			// Get the section name
			sectionName := *param.SectionName

			// Add the parameter to the appropriate section
			groupedParams[sectionName] = append(groupedParams[sectionName], param)
		}
	}

	return groupedParams
}
