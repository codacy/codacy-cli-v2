package pylint

import (
	"codacy/cli-v2/tools/types"
	"fmt"
	"log"
	"strings"
)

// GeneratePylintRC generates a pylintrc file content with the specified patterns enabled
func GeneratePylintRC(config types.ToolConfiguration) string {
	var rcContent strings.Builder

	// Write header
	rcContent.WriteString("[MASTER]\n")
	rcContent.WriteString("ignore=CVS\n")
	rcContent.WriteString("persistent=yes\n")
	rcContent.WriteString("load-plugins=\n\n")

	// Disable all patterns by default
	rcContent.WriteString("[MESSAGES CONTROL]\n")
	rcContent.WriteString("disable=all\n")

	// Collect all enabled pattern IDs and their parameters
	var enabledPatternsIds []string
	patternParams := make(map[string]map[string]string)

	// Only process patterns if the tool is enabled
	if config.IsEnabled {
		for _, pattern := range config.Patterns {
			patternID := extractPatternId(pattern.InternalId)
			enabledPatternsIds = append(enabledPatternsIds, patternID)

			// Store parameters for this pattern
			params := make(map[string]string)
			for _, param := range pattern.Parameters {
				params[param.Name] = param.Value
			}
			patternParams[patternID] = params
		}
	}

	// Write all enabled patterns in a single line
	if len(enabledPatternsIds) > 0 {
		rcContent.WriteString(fmt.Sprintf("enable=%s\n", strings.Join(enabledPatternsIds, ",")))
	}
	rcContent.WriteString("\n")

	// Group parameters by section
	groupedParams := groupParametersBySection(config)

	// Write parameters for each section
	for sectionName, params := range groupedParams {
		rcContent.WriteString(fmt.Sprintf("[%s]\n", sectionName))
		for _, param := range params {
			value := param.Value
			if value == "" {
				// If no value is set, use default from PatternDefaultParameters
				if defaultParams, ok := PatternDefaultParameters[extractPatternId(param.Name)]; ok {
					for _, defaultParam := range defaultParams {
						if defaultParam.Name == param.Name {
							value = defaultParam.Value
							break
						}
					}
				}
			}
			rcContent.WriteString(fmt.Sprintf("%s=%s\n", param.Name, value))
		}
		rcContent.WriteString("\n")
	}

	return rcContent.String()
}

// groupParametersBySection groups parameters by their section name
func groupParametersBySection(config types.ToolConfiguration) map[string][]types.PylintPatternParameterConfiguration {
	// Initialize the result map
	groupedParams := make(map[string][]types.PylintPatternParameterConfiguration)

	// Iterate through each pattern
	for _, pattern := range config.Patterns {
		// Get parameters to process - either from the pattern or from defaults
		if len(pattern.Parameters) == 0 {
			// If no parameters, check defaults
			patternID := extractPatternId(pattern.InternalId)
			pattern.Parameters = PatternDefaultParameters[patternID]
		}

		// Convert pattern parameters to PylintPatternParameterConfiguration
		parameters := make([]types.PylintPatternParameterConfiguration, len(pattern.Parameters))
		for i, param := range pattern.Parameters {
			parameters[i] = types.PylintPatternParameterConfiguration{
				Name:        param.Name,
				Value:       param.Value,
				SectionName: GetParameterSection(param.Name),
			}
		}

		// Process all parameters (either from pattern or defaults)
		for _, param := range parameters {
			// Skip parameters without a section name
			if param.SectionName == nil {
				log.Printf("Parameter %s has no section name", param.Name)
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

// extractPatternId returns the part of the pattern ID after the underscore
// For example: "PyLintPython3_C0301" -> "C0301"
func extractPatternId(fullID string) string {
	parts := strings.Split(fullID, "_")
	if len(parts) > 1 {
		return parts[1]
	}
	return fullID
}
